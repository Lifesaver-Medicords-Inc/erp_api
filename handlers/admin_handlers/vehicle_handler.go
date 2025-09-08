package adminhandlers

import (
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
	"github.com/pierceperado/smpc/utils"
)

type VehicleHandler struct {
	VehicleService *adminservices.VehicleService
}

func NewVehicleHandler(service *adminservices.VehicleService) *VehicleHandler {
	return &VehicleHandler{
		VehicleService: service,
	}
}

func (v *VehicleHandler) CreateVehicle(c *fiber.Ctx) error {
	var body models.VehicleModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := v.VehicleService.CreateVehicle(body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (v *VehicleHandler) GetVehicle(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	position, status, err := v.VehicleService.GetVehicle(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, position)
}

func (v *VehicleHandler) GetVehicles(c *fiber.Ctx) error {
	id := c.Query("id")
	warehouseId := c.Query("warehouseId")
	vehicleType := c.Query("type")
	model := c.Query("model")
	statusQuery := c.Query("status")

	conditions := make(map[string]interface{})

	idNum, _ := strconv.Atoi(id)

	warehouseIdNum, _ := strconv.Atoi(warehouseId)

	if idNum != 0 {
		conditions["id"] = id
	}

	if warehouseIdNum != 0 {
		conditions["ware_house_id"] = warehouseId
	}

	if vehicleType != "" {
		conditions["type"] = vehicleType
	}

	if model != "" {
		conditions["model"] = model
	}

	if statusQuery != "" {
		conditions["status"] = statusQuery
	}

	fmt.Println(warehouseId)

	vehicles, status, err := v.VehicleService.GetVehicles(conditions)

	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, vehicles)
}

func (v *VehicleHandler) UpdateVehicle(c *fiber.Ctx) error {

	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body models.VehicleModel
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	conditions := map[string]interface{}{
		"id": idNum,
	}

	data, status, err := v.VehicleService.UpdateVehicle(body, conditions, at)
	if err != nil {

		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func (v *VehicleHandler) DeleteVehicle(c *fiber.Ctx) error {

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

	data, status, err := v.VehicleService.DeleteVehicle(conditions, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
