package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/NeoPlays/simple-vod/backend/internal/config"
	"github.com/NeoPlays/simple-vod/backend/internal/db"
	"github.com/NeoPlays/simple-vod/backend/internal/ffmpeg"
)



func ReadVideoDirectory() ([]db.Video, error) {
	videoDir := config.GetVideoDirectory()
	
	dirExists, err := os.Stat(videoDir)
	if err != nil {
		// if the error is that the directory does not exist, we can create it
		if errors.Is(err, os.ErrNotExist) {
			err = os.MkdirAll(videoDir, 0755)
			if err != nil {
				return nil, err
			}
			return []db.Video{}, nil
		}
	}

	if !dirExists.IsDir() {
		return nil, errors.New("video directory path is not a directory")
	}

	files, err := os.ReadDir(videoDir)
	if err != nil {
		return nil, err
	}

	videos := make([]db.Video, 0, len(files))
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(strings.ToLower(file.Name()), ".mp4") {
			videos = append(videos, db.Video{ID: -1, Name: file.Name()})
		}
	}

	return videos, nil
}

func (h *Handler) ListVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	videos, err := db.ListVideos(r.Context(), h.DB)
	if err != nil {
		log.Println("Error listing videos:", err)
		http.Error(w, "Failed to list videos", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(videos)
}

func (h *Handler) StreamVideo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/videos/")
	if id == "" {
		http.Error(w, "missing video id", http.StatusBadRequest)
		return
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		http.Error(w, "invalid video id", http.StatusBadRequest)
		return
	}

	v, err := db.GetVideoByID(r.Context(), h.DB, intID)
	if err != nil {
		if errors.Is(err, db.ErrNotFound) {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	videoDir := filepath.Clean(config.GetVideoDirectory())
	fullPath := filepath.Join(videoDir, v.Name)
	if !strings.HasPrefix(fullPath, videoDir+string(os.PathSeparator)) {
		http.Error(w, "forbidden", http.StatusForbidden)
		return
	}

	f, err := os.Open(fullPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			http.Error(w, "video not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "video/mp4")

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}

func (h *Handler) SyncVideos(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	newVideos, err := ReadVideoDirectory()
	if err != nil {
		log.Println("Error reading video directory:", err)
		http.Error(w, "Failed to read video directory", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()

	onDisk := make(map[string]struct{}, len(newVideos))
	for _, v := range newVideos {
		onDisk[v.Name] = struct{}{}
	}

	createdCount := 0
	for _, v := range newVideos {
		created, err := db.CreateVideoIfMissing(ctx, h.DB, v.Name)
		if err != nil {
			log.Println("Error inserting video:", err)
			http.Error(w, "Failed to sync videos", http.StatusInternalServerError)
			return
		}
		if created {
			createdCount++
		}
	}

	existing, err := db.ListVideos(ctx, h.DB)
	if err != nil {
		log.Println("Error listing videos:", err)
		http.Error(w, "Failed to sync videos", http.StatusInternalServerError)
		return
	}

	deletedCount := 0
	for _, v := range existing {
		if _, found := onDisk[v.Name]; !found {
			if err := db.DeleteVideo(ctx, h.DB, v.ID); err != nil {
				log.Println("Error deleting video:", err)
				http.Error(w, "Failed to sync videos", http.StatusInternalServerError)
				return
			}
			deletedCount++
		}
	}

	w.WriteHeader(http.StatusOK)
	log.Printf(`{"created":%d,"deleted":%d,"seen":%d}`, createdCount, deletedCount, len(newVideos))

	go generateThumbnails(existing)
}

func generateThumbnails(videos []db.Video) {
	bin, err := ffmpeg.Ensure(config.GetBinDir())
	if err != nil {
		log.Println("thumbnail generation unavailable:", err)
		return
	}
	videoDir := config.GetVideoDirectory()
	thumbDir := filepath.Join(videoDir, "thumbs")
	for _, v := range videos {
		thumbPath := filepath.Join(thumbDir, ffmpeg.ThumbName(v.Name))
		if _, err := os.Stat(thumbPath); err == nil {
			continue
		}
		videoPath := filepath.Join(videoDir, v.Name)
		if err := ffmpeg.Thumbnail(bin, videoPath, thumbPath); err != nil {
			log.Printf("thumbnail failed for %s: %v", v.Name, err)
		}
	}
}

func (h *Handler) ServeThumb(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	v, err := db.GetVideoByID(r.Context(), h.DB, id)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	videoDir := config.GetVideoDirectory()
	thumbPath := filepath.Join(videoDir, "thumbs", ffmpeg.ThumbName(v.Name))
	f, err := os.Open(thumbPath)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	defer f.Close()

	stat, _ := f.Stat()
	w.Header().Set("Content-Type", "image/jpeg")
	http.ServeContent(w, r, stat.Name(), stat.ModTime(), f)
}
