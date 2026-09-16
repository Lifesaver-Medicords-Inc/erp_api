package public_handlers

import (
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

// resolveInFilesDir returns the absolute path p names inside services.FilesDir(), or
// false when it names anything else. Both spellings in use are accepted: a path as
// stored ("./files/1757...docx", "./files/receipts/..."), and a name relative to the
// folder, which is what /upload hands back. Directories are refused - os.Remove
// deletes an empty one.
func resolveInFilesDir(p string) (string, bool) {
	root, err := filepath.Abs(services.FilesDir())
	if err != nil {
		return "", false
	}

	candidates := []string{p}
	if !filepath.IsAbs(p) && filepath.VolumeName(p) == "" {
		candidates = append(candidates, filepath.Join(root, p))
	}

	for _, candidate := range candidates {
		abs, err := filepath.Abs(candidate)
		if err != nil {
			continue
		}

		rel, err := filepath.Rel(root, abs)
		if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}

		if info, err := os.Lstat(abs); err == nil && info.IsDir() {
			return "", false
		}

		return abs, true
	}

	return "", false
}

func DeleteFile(c *fiber.Ctx) error {
	type Request struct {
		Path string `json:"path"`
	}

	var request Request
	if err := c.BodyParser(&request); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	// An empty path was always a no-op success (services.DeleteFile); kept that way.
	if strings.TrimSpace(request.Path) == "" {
		return utils.RespondSuccess(c, "file deleted")
	}

	// The path used to go straight to os.Remove, so a body naming "../.env" or an
	// absolute path removed that file instead - anything the API process could
	// write. Checked here rather than in services.DeleteFile: the internal callers
	// pass paths the server built itself, not ones a request supplied.
	path, ok := resolveInFilesDir(request.Path)
	if !ok {
		return utils.RespondError(c, fiber.StatusBadRequest, "file path must be inside the files folder")
	}

	// Generic messages: os.Remove's own error now carries the absolute server path.
	if err := services.DeleteFile(path); err != nil {
		if os.IsNotExist(err) {
			return utils.RespondError(c, fiber.StatusNotFound, "file not found")
		}
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed deleting file")
	}

	return utils.RespondSuccess(c, "file deleted")
}

func ViewFile(c *fiber.Ctx) error {
	fileName := c.Params("filename")
	filePath := filepath.Join(services.FilesDir(), fileName)

	file, err := os.Open(filePath)
	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("File not found")
	}

	mimeType := mime.TypeByExtension(filepath.Ext(fileName))
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	c.Set("Content-Type", mimeType)
	c.Set("Content-Disposition", "inline; filename="+fileName)

	return c.SendStream(struct {
		io.Reader
		io.Closer
	}{
		Reader: file,
		Closer: file,
	})
}
