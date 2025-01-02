package internal

import (
	"os"

	"github.com/rs/zerolog/log"
)

func saveFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error().
				Err(err).
				Str("file", path).
				Msg("Failed to close file")
		}
	}(file)

	_, err = file.Write(data)
	if err != nil {
		log.Error().
			Err(err).
			Str("file", path).
			Msg("Failed to write data to file")
		return err
	}

	log.Info().
		Str("file", path).
		Msg("File saved successfully")
	return nil
}

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
