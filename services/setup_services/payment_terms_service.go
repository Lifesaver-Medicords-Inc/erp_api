package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPaymentTerms() ([]models.PaymentTerms, error) {
	var paymentTerms []models.PaymentTerms

	if err := services.DbGet(&paymentTerms, nil); err != nil {
		return paymentTerms, err
	}

	return paymentTerms, nil
}

func GetPaymentTerm(id int) (models.PaymentTerms, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var paymentTerms models.PaymentTerms

	if err := services.DbGet(&paymentTerms, conditions); err != nil {
		return paymentTerms, err
	}

	return paymentTerms, nil
}

func CreatePaymentTerms(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbInsert(tx, &body); err != nil {

		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record")
		}
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func UpdatePaymentTerms(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbUpdate(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: models.At{}}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}

func DeletePaymentTerms(c *fiber.Ctx, tx *gorm.DB) error {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return err
	}

	if err := services.DbDelete(tx, &body, nil); err != nil {
		return err
	}

	// at, ok := c.Locals("at").(models.At)
	// if !ok {
	// 	return errors.New("error AT data")
	// }

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: models.At{}}

	// atdata := models.BrandAt{
	// 	RefId: body.ID,
	// 	Code:  body.Code,
	// 	Name:  body.Name,
	// 	At:    at,
	// }

	if err := services.DbInsert(tx, &atdata); err != nil {
		return err
	}

	return nil
}
