package filehandler

import (
	"fmt"
	"path/filepath"

	"github.com/rs/zerolog/log"
)

type Organizer struct {
	BaseDir string
}

func NewOrganizer(baseDir string) *Organizer {
	return &Organizer{BaseDir: baseDir}
}

func (o *Organizer) OrganizeAndSave(data []byte, mimeType string, chatID, mediaID int) (string, error) {
	extension := getFileExtension(mimeType)
	dirPath := filepath.Join(o.BaseDir, fmt.Sprintf("%d", chatID), mimeType)
	fileName := fmt.Sprintf("%d%s", mediaID, extension)
	fullPath := filepath.Join(dirPath, fileName)

	if err := ensureDir(dirPath); err != nil {
		return "", err
	}

	if err := saveFile(fullPath, data); err != nil {
		return "", err
	}

	log.Info().
		Str("path", fullPath).
		Msg("File organized and saved")
	return fullPath, nil
}
