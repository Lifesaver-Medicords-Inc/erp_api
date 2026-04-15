package adminhandlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type VehicleFileHandler struct {
	VehicleFileService *adminservices.VehicleFileService
}

func NewVehicleFileHandler(service *adminservices.VehicleFileService) *VehicleFileHandler {
	return &VehicleFileHandler{
		VehicleFileService: service,
	}
}

func (h *VehicleFileHandler) DownloadFileHandler(c *fiber.Ctx) error {
	filePathParam := c.Query("path")

	if filePathParam == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "failed: file path missing")
	}

	// Replace Windows slashes with Unix style
	normalizedPath := strings.ReplaceAll(filePathParam, "\\", "/")

	// Clean the path to prevent directory traversal
	cleanPath := filepath.Clean(normalizedPath)

	// Join with base folder
	fullPath := filepath.Join(".", cleanPath)

	// Check path safety
	basePath := filepath.Clean("./files") + string(filepath.Separator)
	if !strings.HasPrefix(filepath.Clean(fullPath), basePath) {
		return c.Status(fiber.StatusForbidden).SendString("Access denied")
	}

	// File existence check
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return c.Status(fiber.StatusNotFound).SendString("File not found")
		}
		return c.Status(fiber.StatusInternalServerError).SendString("Internal server error")
	}

	// Set headers to force file download
	filename := filepath.Base(fullPath)
	c.Set("Content-Disposition", "attachment; filename="+filename)
	c.Set("Content-Type", "application/octet-stream")

	return c.SendFile(fullPath)
}

func (h *VehicleFileHandler) UploadVehicleFileHandler(c *fiber.Ctx) error {
	fileHeader, err := c.FormFile("file")
	if err != nil {
		fmt.Println("FormFile error:", err)
		return utils.RespondError(c, fiber.StatusBadRequest, "fialed open file")
	}

	file, err := fileHeader.Open()
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "failed to open file header")
	}
	defer file.Close()

	vehicleIdStr := c.FormValue("vehicle-id")
	vehicleId, err := strconv.Atoi(vehicleIdStr)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "failed to open file")
	}

	record, status, err := h.VehicleFileService.SaveUploadedFileService(file, fileHeader, vehicleId)

	if err != nil {
		utils.RespondError(c, status, err.Error())
	}

	defer file.Close()

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.VehicleFileService.SaveVehicleFileService(record, at)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return utils.RespondSuccess(c, data)
}

func (h *VehicleFileHandler) GetVehicleFileHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	position, status, err := h.VehicleFileService.GetVehicleFileService(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, position)
}

func (h *VehicleFileHandler) GetVehicleFilesHandler(c *fiber.Ctx) error {
	id := c.Query("id")
	vehicleIdStr := c.Query("vehicle-id")
	fileName := c.Query("file-name")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)
	vehicleIdNum, _ := strconv.Atoi(vehicleIdStr)

	if idNum != 0 {
		conditions["id"] = id
	}

	if vehicleIdNum != 0 {
		conditions["vehicle_id"] = vehicleIdNum
	}

	if fileName != "" {
		conditions["file_name"] = fileName
	}

	vehicles, status, err := h.VehicleFileService.GetVehicleFilesService(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, vehicles)
}

func (h *VehicleFileHandler) DeleteVehicleFileHandler(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.VehicleFileService.RemoveVehicleFileService(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
