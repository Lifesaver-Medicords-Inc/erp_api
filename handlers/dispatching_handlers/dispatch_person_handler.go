package dispatching_handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	dispatching_models "github.com/pierceperado/smpc/models/dispatching_model"
	dispatching_services "github.com/pierceperado/smpc/services/dispatching_service"
	"github.com/pierceperado/smpc/utils"
)

// Dispatch people - the drivers and helpers who go out on deliveries (§13.3). They are
// not system users and do not log in; see the model for why this is separate from both
// the user tables and HRIS (§15).
type DispatchPersonHandler struct {
	Service *dispatching_services.DispatchPersonService
}

func NewDispatchPersonHandler(service *dispatching_services.DispatchPersonService) *DispatchPersonHandler {
	return &DispatchPersonHandler{Service: service}
}

// GET /dispatch-people            - everyone, including inactive (Setup needs to see them)
// GET /dispatch-people?active=1   - only people who can still be assigned (schedule pickers)
// GET /dispatch-people?role=DRIVER
func (h *DispatchPersonHandler) GetDispatchPeopleHandler(c *fiber.Ctx) error {
	conditions := map[string]interface{}{}

	if c.Query("active") == "1" || strings.EqualFold(c.Query("active"), "true") {
		conditions["is_active"] = true
	}

	if role := strings.ToUpper(strings.TrimSpace(c.Query("role"))); role != "" {
		if !dispatching_services.IsValidDispatchRole(role) {
			return utils.RespondError(c, fiber.StatusBadRequest, "role must be DRIVER or HELPER")
		}
		conditions["role"] = role
	}

	people, status, err := h.Service.GetDispatchPeople(conditions)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, people)
}

// GET /dispatch-people/:id
func (h *DispatchPersonHandler) GetDispatchPersonHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	person, status, err := h.Service.GetDispatchPerson(map[string]interface{}{"id": id})
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, person)
}

// POST /dispatch-people
func (h *DispatchPersonHandler) CreateDispatchPersonHandler(c *fiber.Ctx) error {
	var body dispatching_models.DispatchPerson
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.CreateDispatchPerson(&body, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// PUT /dispatch-people/:id
func (h *DispatchPersonHandler) UpdateDispatchPersonHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	var body dispatching_models.DispatchPerson
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.UpdateDispatchPerson(&body, map[string]interface{}{"id": id}, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// DELETE /dispatch-people/:id - deactivates, never removes. Past schedules keep the
// person's name; only the pickers stop offering them (§4.4.1's inactive-warehouse rule).
func (h *DispatchPersonHandler) DeactivateDispatchPersonHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "Invalid ID parameter")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	data, status, err := h.Service.DeactivateDispatchPerson(map[string]interface{}{"id": id}, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}
