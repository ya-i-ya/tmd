package filehandler

import (
	"fmt"
	"github.com/gotd/td/tg"
	"os"
	"path/filepath"
	"strings"

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

func GetMimeType(media tg.MessageMediaClass) (string, error) {
	switch m := media.(type) {
	case *tg.MessageMediaPhoto:
		return "image/jpeg", nil
	case *tg.MessageMediaDocument:
		docObj, ok := m.Document.(*tg.Document)
		if !ok || docObj == nil {
			return "", fmt.Errorf("document is not a *tg.Document or is nil")
		}
		if docObj.MimeType != "" {
			return docObj.MimeType, nil
		}
		return "application/octet-stream", nil
	default:
		return "", fmt.Errorf("unsupported media type: %T", m)
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

func BuildObjectName(dialogName, mimeType string, messageID int) string {
	mimeDir := firstSegmentOfMime(mimeType)
	ext := getFileExtension(mimeType)
	fileName := fmt.Sprintf("%d%s", messageID, ext)
	return filepath.ToSlash(filepath.Join(dialogName, mimeDir, fileName))
}

func firstSegmentOfMime(mime string) string {
	if idx := strings.IndexRune(mime, '/'); idx > 0 {
		return mime[:idx]
	}
	return mime
}
