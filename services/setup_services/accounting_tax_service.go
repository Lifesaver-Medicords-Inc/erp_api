package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type BodyRequest struct {
	accounting_models.Tax
}

func CreateTaxSetup(c *fiber.Ctx, tx *gorm.DB) (BodyRequest, int, error) {

	var body BodyRequest
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating tax code")
		}
		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := accounting_models.TaxAt{RefId: body.ID, TaxContent: body.TaxContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// key := services.GetKey(accounting_models.ChartOfAccountViewList{}, nil)
	//services.InvalidateCache(key)
	return body, 0, nil
}
