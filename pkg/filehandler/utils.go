package filehandler

import (
	"os"

	"github.com/rs/zerolog/log"
)

func getFileExtension(mimeType string) string {
	switch mimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/pdf":
		return ".pdf"
	case "video/mp4":
		return ".mp4"
	case "video/mpeg":
		return ".mpeg"
	case "audio/mpeg":
		return ".mp3"
	case "application/zip":
		return ".zip"
	case "audio/ogg":
		return ".ogg"
	default:
		return ".dat"
	}
}

func ensureDir(path string) error {
	err := os.MkdirAll(path, os.ModePerm)
	if err != nil {
		log.Error().
			Err(err).
			Str("path", path).
			Msg("Failed to create directory")
		return err
	}
	log.Debug().
		Str("path", path).
		Msg("Directory ensured")
	return nil
}
