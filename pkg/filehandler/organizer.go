package filehandler

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/rs/zerolog/log"
)

type Organizer struct {
	BaseDir string
}

func NewOrganizer(baseDir string) *Organizer {
	return &Organizer{BaseDir: baseDir}
}

func (o *Organizer) getDirPath(mimeType string, dialogName string) (string, error) {
	mimeDir := strings.TrimRight(mimeType, "/")
	dirPath := filepath.Join(o.BaseDir, fmt.Sprintf("%s", dialogName), mimeDir)
	if err := ensureDir(dirPath); err != nil {
		return "", err
	}

	log.Info().
		Str("path", dirPath).
		Msg("Returned dir path")
	return dirPath, nil
}

func (o *Organizer) getFilePath(mimeType, dialogName string) (string, error) {
	extension := getFileExtension(mimeType)
	dirPath, err := o.getDirPath(dialogName, mimeType)
	if err != nil {
		return "", err
	}
	fileName := fmt.Sprintf("%s%s", dialogName, extension)
	fullPath := filepath.Join(dirPath, fileName)

	log.Info().
		Str("path", fullPath).
		Msg("Returned file path")

	return fullPath, nil
}
