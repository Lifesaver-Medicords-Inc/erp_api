package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

func UploadFile(filestr string) (string, error) {
	var path string

	file, err := base64.StdEncoding.DecodeString(filestr)
	if err != nil {
		return path, errors.New("failed decoding data")
	}

	if err := os.MkdirAll("./files", os.ModePerm); err != nil {
		return path, errors.New("failed creating folder")
	}

	name := time.Now().UnixNano()
	mime := mimetype.Detect(file)
	extension := mime.Extension()

	path = fmt.Sprintf("./files/%d%v", name, extension)

	if err := os.WriteFile(path, file, 0644); err != nil {
		return path, errors.New("failed saving file")
	}

	return path, nil
}

func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil // No file to delete
	}
	return os.Remove(filePath)
}
