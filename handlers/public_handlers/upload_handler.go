package public_handlers

import (
	"encoding/base64"
	"fmt"
	"os"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/utils"
)

func ImageUpload(c *fiber.Ctx) error {
	type Request struct {
		File string `json:"file"`
	}

	var request Request
	if err := c.BodyParser(&request); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	file, err := base64.StdEncoding.DecodeString(request.File)
	if err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed decoding data")
	}

	if err := os.MkdirAll("./files", os.ModePerm); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed creating folder")
	}

	fileName := time.Now().Unix()
	mimeType := mimetype.Detect(file)
	fileExtension := mimeType.Extension()

	path := fmt.Sprintf("./files/%d%v", fileName, fileExtension)

	if err := os.WriteFile(path, file, 0644); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "failed saving file")
	}

	return utils.RespondSuccess(c, file)
}
