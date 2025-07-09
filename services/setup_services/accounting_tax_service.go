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
	TaxDetails []accounting_models.TaxDetails `json:"tax_details"`
}

func GetTaxSetup(conditions map[string]interface{}) (interface{}, int, error) {

	type Response struct {
		Tax        []accounting_models.Tax        `json:"tax"`
		TaxDetails []accounting_models.TaxDetails `json:"tax_details"`
		TaxView    []accounting_models.TaxView    `json:"tax_view"`
	}

	var response Response

	if err := services.DbGet(&response.Tax, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get tax setup")
	}

	if err := services.DbGet(&response.TaxDetails, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get tax details setup")
	}

	if err := services.DbGet(&response.TaxView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get tax view setup")
	}
	return response, 0, nil
}

func CreateTaxSetup(c *fiber.Ctx, tx *gorm.DB) (BodyRequest, int, error) {

	var body BodyRequest
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.Tax); err != nil {
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

	if err := CreateTaxDetails(tx, body.TaxDetails, body.ID, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	key := services.GetKey(accounting_models.TaxView{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}

// Other Services
func CreateTaxDetails(tx *gorm.DB, child []accounting_models.TaxDetails, parentId uint, at models.At) error {

	for _, v := range child {
		if err := CreateTaxDetail(tx, v, parentId, at); err != nil {
			return err
		}
	}

	return nil
}

func CreateTaxDetail(tx *gorm.DB, child accounting_models.TaxDetails, parentId uint, at models.At) error {

	child.TaxDetailsContent.TaxCodeId = parentId

	if err := services.DbInsert(tx, &child); err != nil {
		return errors.New("failed to insert tax details table")
	}
	childAt := accounting_models.TaxDetailsAt{
		RefId:             child.ID,
		TaxDetailsContent: child.TaxDetailsContent,
		At:                at,
	}
	if err := services.DbInsert(tx, &childAt); err != nil {
		return errors.New("failed to insert tax details at table")
	}

	return nil
}
