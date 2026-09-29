package setup_handlers

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"

	"github.com/pierceperado/smpc/services/setup_services"
	"github.com/pierceperado/smpc/utils"
)

func GetSetupItemBoms(c *fiber.Ctx) error {
	data, status, err := setup_services.GetSetupItemBoms(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetBomsPaged serves the BOM screen a page at a time: ?after=, ?before= or ?at= (a BOM id),
// none for the first page. Same bom_head/bom_details shape as GET /setup/bom.
func GetBomsPaged(c *fiber.Ctx) error {
	after, _ := strconv.Atoi(c.Query("after", "0"))
	before, _ := strconv.Atoi(c.Query("before", "0"))
	at, _ := strconv.Atoi(c.Query("at", "0"))

	data, pagination, status, err := setup_services.GetBomsPaged(after, before, at)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

// GetBomTree serves the quotation screens one BOM (or a few) with every sub-BOM under them:
// ?ids=12,40 - at most 500 ids a request.
func GetBomTree(c *fiber.Ctx) error {
	var ids []int
	for _, part := range strings.Split(c.Query("ids"), ",") {
		if id, err := strconv.Atoi(strings.TrimSpace(part)); err == nil && id > 0 {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return utils.RespondError(c, fiber.StatusBadRequest, "no bom ids given")
	}
	if len(ids) > setup_services.BomTreeMaxIDs {
		return utils.RespondError(c, fiber.StatusBadRequest, "too many bom ids in one request")
	}

	data, status, err := setup_services.GetBomTree(ids)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

// GetBomsSearch serves the BOM screen's Search: matching heads, 20 to a page. ?search=, ?page=
// (1-based).
func GetBomsSearch(c *fiber.Ctx) error {
	search := c.Query("search")
	page, _ := strconv.Atoi(c.Query("page", "1"))

	data, pagination, status, err := setup_services.GetBomsSearch(search, page)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data, pagination)
}

func GetSetupItemBomss(c *fiber.Ctx) error {
	data, status, err := setup_services.GetSetupItemBomss(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetSetupItemBom(c *fiber.Ctx) error {
	idParam := c.Params("id")
	idNum, err := strconv.Atoi(idParam)
	if err != nil {
		return utils.RespondError(c, fiber.StatusBadRequest, err.Error())
	}

	data, status, err := setup_services.GetSetupItemBom(idNum)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetBomItemList(c *fiber.Ctx) error {
	data, status, err := setup_services.GetBomItemList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetAllBomItemList(c *fiber.Ctx) error {
	data, status, err := setup_services.GetAllBomItemList(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetBomParentDetail(c *fiber.Ctx) error {
	data, status, err := setup_services.GetBomParentDetail(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func GetBomChildDetail(c *fiber.Ctx) error {
	data, status, err := setup_services.GetBomChildDetail(nil)
	if err != nil {
		return utils.RespondError(c, status, err.Error())
	}

	return utils.RespondSuccess(c, data)
}

func CreateSetupItemBom(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.CreateSetupItemBom(c, tx)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func UpdateSetupItemBom(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.UpdateSetupItemBom(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}

func DeleteSetupItemBom(c *fiber.Ctx) error {
	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to start transaction")
	}

	data, status, err := setup_services.DeleteSetupItemBom(c, tx, nil)
	if err != nil {
		tx.Rollback()
		return utils.RespondError(c, status, err.Error())
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return utils.RespondError(c, fiber.StatusInternalServerError, "Failed to commit transaction")
	}

	return utils.RespondSuccess(c, data)
}
