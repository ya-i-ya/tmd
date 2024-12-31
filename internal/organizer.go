package internal

import (
	"github.com/gotd/td/tg"
	"os"
)

func saveFile(path string, data []byte) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
		}
	}(file)

	_, err = file.Write(data)
	return err
}

func getFileExtension(document *tg.Document) string {
	switch document.MimeType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "application/pdf":
		return ".pdf"
	// etc.
	default:
		return ".dat"
	}
}
