package adminhandlers

import (
	"fmt"
	"strconv"

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

func (h *VehicleFileHandler) UploadHandler(c *fiber.Ctx) error {

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

	record, status, err := h.VehicleFileService.SaveUploadedFile(file, fileHeader, vehicleId)

	if err != nil {
		utils.RespondError(c, status, err.Error())
	}

	defer file.Close()

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.VehicleFileService.SaveVehicleFile(record, at)

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
