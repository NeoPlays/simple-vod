package ffmpeg

import (
	"archive/tar"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/ulikunitz/xz"
)

var (
	mu     sync.Mutex
	cached string
)

// Ensure returns the path to the ffmpeg binary.
// On Linux it downloads a static build into binDir if needed.
// On other platforms it falls back to whatever ffmpeg is on $PATH.
func Ensure(binDir string) (string, error) {
	mu.Lock()
	defer mu.Unlock()

	if cached != "" {
		return cached, nil
	}

	if runtime.GOOS != "linux" {
		path, err := exec.LookPath("ffmpeg")
		if err != nil {
			return "", fmt.Errorf("ffmpeg not found on PATH (non-Linux host)")
		}
		cached = path
		return path, nil
	}

	dest := filepath.Join(binDir, "ffmpeg")
	if _, err := os.Stat(dest); err == nil {
		cached = dest
		return dest, nil
	}

	log.Println("ffmpeg not found, downloading static build…")
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return "", fmt.Errorf("create bin dir: %w", err)
	}
	if err := download(dest); err != nil {
		return "", fmt.Errorf("download ffmpeg: %w", err)
	}
	cached = dest
	log.Println("ffmpeg ready at", dest)
	return dest, nil
}

func download(dest string) error {
	url, err := staticURL()
	if err != nil {
		return err
	}
	log.Println("downloading", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d fetching ffmpeg", resp.StatusCode)
	}

	xr, err := xz.NewReader(resp.Body)
	if err != nil {
		return fmt.Errorf("xz reader: %w", err)
	}
	tr := tar.NewReader(xr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if hdr.Typeflag != tar.TypeReg || filepath.Base(hdr.Name) != "ffmpeg" {
			continue
		}
		// skip ffprobe/ffplay entries whose base happens to match
		if strings.Contains(filepath.Base(hdr.Name), "probe") || strings.Contains(filepath.Base(hdr.Name), "play") {
			continue
		}
		f, err := os.OpenFile(dest, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(f, tr)
		f.Close()
		if copyErr != nil {
			os.Remove(dest)
			return copyErr
		}
		return nil
	}
	return fmt.Errorf("ffmpeg binary not found in archive")
}

func staticURL() (string, error) {
	arch := runtime.GOARCH
	switch arch {
	case "amd64", "arm64":
	case "arm":
		arch = "armhf"
	default:
		return "", fmt.Errorf("unsupported arch: %s", arch)
	}
	return fmt.Sprintf(
		"https://johnvansickle.com/ffmpeg/releases/ffmpeg-release-%s-static.tar.xz", arch,
	), nil
}

// Thumbnail extracts a single frame at 5 s and saves it as a JPEG.
func Thumbnail(ffmpegPath, videoPath, thumbPath string) error {
	if err := os.MkdirAll(filepath.Dir(thumbPath), 0755); err != nil {
		return err
	}
	out, err := exec.Command(ffmpegPath,
		"-ss", "10",
		"-i", videoPath,
		"-vframes", "1",
		"-q:v", "3",
		"-vf", "scale=480:-1",
		"-y",
		thumbPath,
	).CombinedOutput()
	if err != nil {
		return fmt.Errorf("ffmpeg: %w\n%s", err, out)
	}
	return nil
}

// ThumbName returns the thumbnail filename for a given video filename.
func ThumbName(videoName string) string {
	return strings.TrimSuffix(videoName, filepath.Ext(videoName)) + ".jpg"
}
