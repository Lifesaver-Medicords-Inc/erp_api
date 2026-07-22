package services

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
)

// FilesDir returns the directory uploaded files are saved to and served from. Defaults to
// "./files" (unchanged behavior) - relative to whatever the process's current working
// directory happens to be at the moment it's used, which can differ across how/where the
// API gets launched from (a rebuild that runs from a different output folder, `go run` vs a
// built exe, an IDE debugger's working directory, etc). Files uploaded during one run can
// end up silently orphaned (404 on /vfile/:filename) if a later run resolves "./files" to a
// different location - this is what caused test images uploaded earlier in a session to 404
// after later rebuilds. Set FILES_DIR in .env to an absolute path to pin this down for good.
func FilesDir() string {
	if dir := os.Getenv("FILES_DIR"); dir != "" {
		return dir
	}
	return "./files"
}

func UploadFile(filestr string) (string, error) {
	var filename string

	file, err := base64.StdEncoding.DecodeString(filestr)
	if err != nil {
		return filename, errors.New("failed decoding data")
	}

	dir := FilesDir()

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return filename, errors.New("failed creating folder")
	}

	name := time.Now().UnixNano()
	mime := mimetype.Detect(file)
	extension := mime.Extension()

	filename = fmt.Sprintf("%d%v", name, extension)
	path := filepath.Join(dir, filename)

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
