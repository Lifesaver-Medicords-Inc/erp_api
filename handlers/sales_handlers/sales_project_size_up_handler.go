package sales_handlers

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services/sales_services"
	"github.com/pierceperado/smpc/utils"
)

// Size Up (spec §5.1.4) normally rides along inside the project quote's own save - the
// list belongs to a content row and is written by syncContentChildren as part of
// POST/PUT /sales/projects. These four endpoints address a single row directly, for the
// cases the bulk save cannot serve: editing one candidate without submitting the whole
// quote, and reading a tab's list back on its own (e.g. to populate Final Selection's
// picker, which §5.1.4 limits to what Size Up holds).
//
// A row is always owned by exactly one content row, so sales_project_content_id is
// required on create and is what GET filters by.

func atFromCtx(c *fiber.Ctx) models.At {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		return models.At{}
	}
	return at
}

// GET /api/sales/project_size_up?content_id=123
// Without content_id this returns every row, which is only useful for support work -
// the client always scopes it to one tab's content row.
func GetProjectSizeUps(c *fiber.Ctx) error {
	conditions := map[string]interface{}{}

	if raw := c.Query("content_id"); raw != "" {
		contentId, err := strconv.ParseUint(raw, 10, 64)
		if err != nil {
			return utils.RespondError(c, fiber.StatusBadRequest, "content_id must be a number")
		}
		conditions["sales_project_content_id"] = uint(contentId)
	}

	var data []models.SalesProjectSizeUp
	if err := sales_services.GetSalesProjectSizeUp(&data, conditions); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// POST /api/sales/project_size_up
func CreateProjectSizeUpRow(c *fiber.Ctx) error {
	var body models.SalesProjectSizeUp
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	if body.SalesProjectContentID == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "sales_project_content_id is required")
	}
	if body.ItemID == 0 && body.Model == "" {
		return utils.RespondError(c, fiber.StatusBadRequest, "a size up row needs an item_id or a model")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	// CreateProjectSizeUp zeroes ID itself and stamps the FK from parentId, so a client
	// that echoes back an id it shouldn't have cannot force an IDENTITY_INSERT.
	if err := sales_services.CreateProjectSizeUp(tx, body.SalesProjectContentID, body, atFromCtx(c)); err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	// Re-read rather than returning body: the caller needs the DB-assigned id, and body
	// is a value copy that CreateProjectSizeUp's own copy never wrote back to.
	var saved []models.SalesProjectSizeUp
	if err := sales_services.GetSalesProjectSizeUp(&saved, map[string]interface{}{
		"sales_project_content_id": body.SalesProjectContentID,
	}); err != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	return utils.RespondSuccess(c, saved)
}

// PUT /api/sales/project_size_up
func UpdateProjectSizeUpRow(c *fiber.Ctx) error {
	var body models.SalesProjectSizeUp
	if err := c.BodyParser(&body); err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, "cannot bind request")
	}

	if body.ID == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "id is required")
	}

	existing, err := loadSizeUpRow(body.ID)
	if err != nil {
		return utils.RespondError(c, fiber.StatusNotFound, "size up row not found")
	}

	// The owning content row is not the client's to change - a row moving between tabs
	// would silently rewrite two Size Up lists at once.
	body.SalesProjectContentID = existing.SalesProjectContentID

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	if err := sales_services.UpdateProjectSizeUp(tx, body, atFromCtx(c), map[string]interface{}{
		"id": body.ID,
	}); err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, body)
}

// DELETE /api/sales/project_size_up/:id
func DeleteProjectSizeUpRow(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil || id == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "id must be a number")
	}

	// Read first so the audit row records what was actually removed, not an empty shell.
	existing, err := loadSizeUpRow(uint(id))
	if err != nil {
		return utils.RespondError(c, fiber.StatusNotFound, "size up row not found")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	if err := sales_services.DeleteProjectSizeUp(tx, existing, atFromCtx(c)); err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, existing)
}

func loadSizeUpRow(id uint) (models.SalesProjectSizeUp, error) {
	var row models.SalesProjectSizeUp
	// Straight off initializers.DB, not the cached DbGet - this reads the row so the
	// update/delete can be validated against it, and a stale cache hit here would let a
	// row that no longer exists be "found".
	err := initializers.DB.Where("id = ?", id).First(&row).Error
	return row, err
}
