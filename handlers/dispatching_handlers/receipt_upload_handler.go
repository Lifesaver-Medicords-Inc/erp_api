package dispatching_handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/pierceperado/smpc/utils"
)

// Generic receipt upload — saves a file under ./files/receipts and returns its
// stored path. Used for cost-row receipts (e.g. logistics route costs) where
// the receipt is just a loose path string on the record, not a separate
// DB-tracked file entity.
func UploadReceiptHandler(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "failed to read uploaded file")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "failed to open uploaded file")
	}
	defer file.Close()

	uploadDir := "./files/receipts"
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed to prepare upload directory")
	}

	ext := filepath.Ext(fileHeader.Filename)
	newFileName := uuid.New().String() + ext
	savePath := filepath.Join(uploadDir, newFileName)

	dst, err := os.Create(savePath)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed to save file")
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed to save file")
	}

	buf := make([]byte, 512)
	dst.Seek(0, io.SeekStart)
	n, _ := dst.Read(buf)
	fileType := http.DetectContentType(buf[:n])

	return utils.RespondSuccess(c, fiber.Map{
		"file_path":     savePath,
		"original_name": fileHeader.Filename,
		"type":          fileType,
		"size":          fileHeader.Size,
	})
}
