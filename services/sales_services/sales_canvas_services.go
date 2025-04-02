package sales_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type CanvasBody struct {
	SalesCanvasSheet []models.SalesCanvasSheet
}

func GetSalesCanvasView(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		SalesCanvasSheetView []models.SalesCanvasSheetView `json:"sales_canvas_sheet_view"`
	}

	var response Response

	if err := services.DbGet(&response.SalesCanvasSheetView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting canvas sheet view")
	}

	return response, 0, nil
}

func CreateSalesCanvasSheet(c *fiber.Ctx, tx *gorm.DB) (CanvasBody, int, error) {
	var body CanvasBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesCanvasSheet); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating canvas sheet")
	}

	return body, 0, nil
}
