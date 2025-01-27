package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

func GetPaymentTerms(conditions map[string]interface{}) ([]models.PaymentTerms, int, error) {
	var paymentTerms []models.PaymentTerms

	if err := services.DbGet(&paymentTerms, conditions); err != nil {
		return paymentTerms, fiber.StatusInternalServerError, errors.New("failed getting payment terms")
	}

	return paymentTerms, 0, nil
}

func GetPaymentTerm(id int) (models.PaymentTerms, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var paymentTerms models.PaymentTerms

	if err := services.DbGet(&paymentTerms, conditions); err != nil {
		return paymentTerms, fiber.StatusInternalServerError, errors.New("failed getting payment terms")
	}

	return paymentTerms, 0, nil
}

func CreatePaymentTerms(c *fiber.Ctx, tx *gorm.DB) (models.PaymentTerms, int, error) {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating payment terms")
		}

		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment terms at")
	}

	return body, 0, nil
}

func UpdatePaymentTerms(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PaymentTerms, int, error) {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating payment terms")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment terms at")
	}

	return body, 0, nil
}

func DeletePaymentTerms(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.PaymentTerms, int, error) {
	var body models.PaymentTerms
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting payment terms")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.PaymentTermsAt{RefId: body.ID, Code: body.Code, PaymentTermsContent: models.PaymentTermsContent{Name: body.Name}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating payment terms at")
	}

	return body, 0, nil
}
