package dispatching_handlers

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

type CalendarScheduleHandler struct {
	CalendarScheduleService    *dispatching_services.CalendarScheduleService
	SalesScheduleService       *dispatching_services.SalesCalendarScheduleService
	EngineeringScheduleService *dispatching_services.EngineeringCalendarScheduleService
	LogisticsScheduleService   *dispatching_services.LogisticsCalendarScheduleService
}

func NewCalendarScheduleHandler(service *dispatching_services.CalendarScheduleService) *CalendarScheduleHandler {
	return &CalendarScheduleHandler{
		CalendarScheduleService:    service,
		SalesScheduleService:       dispatching_services.NewSalesCalendarScheduleService(),
		EngineeringScheduleService: dispatching_services.NewEngineeringCalendarScheduleService(),
		LogisticsScheduleService:   dispatching_services.NewLogisticsCalendarScheduleService(),
	}
}

func getAtFromLocals(c *fiber.Ctx) models.At {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		return models.At{}
	}
	return at
}

func parseIDParam(c *fiber.Ctx, paramName string) (int, error) {
	idParam := c.Params(paramName)
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return 0, err
	}
	return idNum, nil
}

func parseQueryInt(c *fiber.Ctx, name string) (int, bool, error) {
	raw := c.Query(name)
	if raw == "" {
		return 0, false, nil
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return 0, true, err
	}
	return n, true, nil
}

func parseQueryBoolLoose(c *fiber.Ctx, name string) (bool, bool) {
	raw := c.Query(name)
	if raw == "" {
		return false, false
	}
	// Preserve prior behavior: any non-empty value other than "true" is treated as false.
	return raw == "true", true
}

func parseDepartmentFromBody(c *fiber.Ctx) (string, error) {
	var dto CalendarScheduleDepartmentDTO
	if err := json.Unmarshal(c.Body(), &dto); err != nil {
		return "", err
	}
	return dto.Department, nil
}

// ✅ GET /calendar-schedules
func (h *CalendarScheduleHandler) GetCalendarSchedulesHandler(c *fiber.Ctx) error {
	department := c.Query("department")
	if department == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "department query parameter is required")
	}

	conditions := make(map[string]interface{})

	if idNum, ok, err := parseQueryInt(c, "id"); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid id query parameter")
	} else if ok && idNum != 0 {
		conditions["id"] = idNum
	}

	conditions["department"] = department

	if isCancelled, ok := parseQueryBoolLoose(c, "is_cancelled"); ok {
		conditions["is_cancelled"] = isCancelled
	}

	switch department {
	case "SALES":
		schedules, status, err := h.SalesScheduleService.GetSalesSchedules(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedules)
	case "ENGINEERING":
		schedules, status, err := h.EngineeringScheduleService.GetEngineeringSchedules(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedules)
	case "LOGISTICS":
		schedules, status, err := h.LogisticsScheduleService.GetLogisticsSchedules(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedules)
	default:
		return utils.RespondError(c, fiber.StatusBadRequest, "Unknown department")
	}
}

// ✅ GET /calendar-schedules/:id
func (h *CalendarScheduleHandler) GetCalendarScheduleHandler(c *fiber.Ctx) error {
	idNum, err := parseIDParam(c, "id")
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	department := c.Query("department")
	if department == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "department query parameter is required")
	}

	conditions := map[string]interface{}{
		"id":         idNum,
		"department": department,
	}

	switch department {
	case "SALES":
		schedule, status, err := h.SalesScheduleService.GetSalesSchedule(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedule)
	case "ENGINEERING":
		schedule, status, err := h.EngineeringScheduleService.GetEngineeringSchedule(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedule)
	case "LOGISTICS":
		schedule, status, err := h.LogisticsScheduleService.GetLogisticsSchedule(conditions)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, schedule)
	default:
		return utils.RespondError(c, fiber.StatusBadRequest, "Unknown department")
	}
}

type CalendarScheduleDepartmentDTO struct {
	Department string `json:"department"`
}

// ✅ POST /calendar-schedules
func (h *CalendarScheduleHandler) CreateCalendarScheduleHandler(c *fiber.Ctx) error {
	department, err := parseDepartmentFromBody(c)
	fmt.Println("request body", department)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at := getAtFromLocals(c)

	// 🔹 Step 2: switch by department
	switch department {

	case "SALES":
		var body dispatching_models.SalesCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Sales schedule body")
		}

		data, status, err := h.SalesScheduleService.CreateSalesSchedule(&body, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "ENGINEERING":
		var body dispatching_models.EngineeringCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Engineering schedule body")
		}

		data, status, err := h.EngineeringScheduleService.CreateEngineeringSchedule(&body, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "LOGISTICS":
		var body dispatching_models.LogisticsCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Logistics schedule body")
		}

		data, status, err := h.LogisticsScheduleService.CreateLogisticsSchedule(&body, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	default:
		return utils.RespondError(c, fiber.StatusBadRequest, "Unknown department")
	}
}

// ✅ PUT /calendar-schedules/:id
func (h *CalendarScheduleHandler) UpdateCalendarScheduleHandler(c *fiber.Ctx) error {
	idNum, err := parseIDParam(c, "id")
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	department, err := parseDepartmentFromBody(c)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at := getAtFromLocals(c)

	conditions := map[string]interface{}{
		"id": idNum,
	}

	switch department {

	case "SALES":
		var body dispatching_models.SalesCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Sales schedule body")
		}

		data, status, err :=
			h.SalesScheduleService.UpdateSalesSchedule(&body, conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "ENGINEERING":
		var body dispatching_models.EngineeringCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Engineering schedule body")
		}

		data, status, err :=
			h.EngineeringScheduleService.UpdateEngineeringSchedule(&body, conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "LOGISTICS":
		var body dispatching_models.LogisticsCalendarScheduleModel
		if err := json.Unmarshal(c.Body(), &body); err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "Invalid Logistics schedule body")
		}

		data, status, err :=
			h.LogisticsScheduleService.UpdateLogisticsSchedule(&body, conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	default:
		return utils.RespondError(c, fiber.StatusBadRequest, "Unknown department")
	}
}
func (h *CalendarScheduleHandler) DeleteSalesScheduleHandler(c *fiber.Ctx) error {
	return h.deleteSchedule(c, "SALES")
}
func (h *CalendarScheduleHandler) DeleteEngineeringScheduleHandler(c *fiber.Ctx) error {
	return h.deleteSchedule(c, "ENGINEERING")
}
func (h *CalendarScheduleHandler) DeleteLogisticsScheduleHandler(c *fiber.Ctx) error {
	return h.deleteSchedule(c, "LOGISTICS")
}

// ✅ DELETE /calendar-schedules/:id
func (h *CalendarScheduleHandler) deleteSchedule(c *fiber.Ctx, department string) error {
	idNum, err := parseIDParam(c, "id")
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	at := getAtFromLocals(c)
	conditions := map[string]interface{}{"id": idNum}

	// only validates and calls the correct service
	switch department {
	case "SALES":
		data, status, err := h.SalesScheduleService.DeleteSalesSchedule(conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "ENGINEERING":
		data, status, err := h.EngineeringScheduleService.DeleteEngineeringSchedule(conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	case "LOGISTICS":
		data, status, err := h.LogisticsScheduleService.DeleteLogisticsSchedule(conditions, at)
		if err != nil {
			return utils.RespondError(c, status, err.Error())
		}
		return utils.RespondSuccess(c, data)

	default:
		return utils.RespondError(c, fiber.StatusBadRequest, "Unknown department")
	}
}
