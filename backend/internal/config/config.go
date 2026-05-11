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

func GetPasswordPepper() string {
	pepper := os.Getenv("PASSWORD_PEPPER")
	if pepper == "" {
		return "default-pepper"
	}
	return pepper
}

func GetFrontendDirectory() string {
	dir := os.Getenv("FRONTEND_DIR")
	if dir == "" {
		return "../frontend"
	}
	return dir
}

func GetBinDir() string {
	dir := os.Getenv("BIN_DIR")
	if dir == "" {
		return "./data/bin"
	}
	return dir
}

func GetDBPath() string {
	path := os.Getenv("DB_PATH")
	if path == "" {
		return "./db/sqlite.db"
	}
	return path
}

func GetPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		return ":8080"
	}
	return ":" + port
}
