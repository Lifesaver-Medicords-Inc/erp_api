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
	var filename string

	file, err := base64.StdEncoding.DecodeString(filestr)
	if err != nil {
		return filename, errors.New("failed decoding data")
	}

	if err := os.MkdirAll("./files", os.ModePerm); err != nil {
		return filename, errors.New("failed creating folder")
	}

	name := time.Now().UnixNano()
	mime := mimetype.Detect(file)
	extension := mime.Extension()

	filename = fmt.Sprintf("%d%v", name, extension)
	path := fmt.Sprintf("./files/%s", filename)

	if err := os.WriteFile(path, file, 0644); err != nil {
		return filename, errors.New("failed saving file")
	}

	return filename, nil
}

func DeleteFile(filePath string) error {
	if filePath == "" {
		return nil
	}
	return os.Remove(filePath)
}
