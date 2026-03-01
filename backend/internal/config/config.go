package config

import (
	"os"
)

func GetVideoDirectory() string {
	dir := os.Getenv("VIDEO_DIR")
	if dir == "" {
		return "./data/videos"
	}
	return dir
}
