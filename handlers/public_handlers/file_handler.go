package public_handlers

import (
	"io"
	"mime"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

func DeleteFile(c *fiber.Ctx) error {
	type Request struct {
		Path string `json:"path"`
	}

	var request Request
	if err := c.BodyParser(&request); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	if err := services.DeleteFile(request.Path); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
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
