package sales_services

import (
	"errors"
	"fmt"

	// fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// test//test
func GetCRMs(conditions map[string]interface{}) ([]models.CRMView, int, error) {
	var CRMs []models.CRMView

	if err := services.DbGet(&CRMs, conditions); err != nil {
		fmt.Println("ERROR:", err.Error())
		return CRMs, fiber.StatusInternalServerError, errors.New("failed getting CRMs")
	}
	// fmt.Println("DATA: ", CRMs)
	return CRMs, 0, nil
}

func GetCRMTable(conditions map[string]interface{}) ([]models.CRM, int, error) {
	var CRMs []models.CRM

	if err := services.DbGet(&CRMs, conditions); err != nil {
		fmt.Println("ERROR:", err.Error())
		return CRMs, fiber.StatusInternalServerError, errors.New("failed getting CRMs")
	}
	// fmt.Println("DATA: ", CRMs)
	return CRMs, 0, nil
}

func GetCRM(id int) (models.CRMView, int, error) {
	conditions := map[string]interface{}{
		"crm_id": id,
	}

	var CRMs models.CRMView

	if err := services.DbGet(&CRMs, conditions); err != nil {
		return CRMs, fiber.StatusInternalServerError, errors.New("failed getting crm")
	}

	return CRMs, 0, nil
}

// test
func CreateCRM(c *fiber.Ctx, tx *gorm.DB) (models.CRM, int, error) {
	var body models.CRM
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("ERROR:", err.Error())
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating crk=m")
		}
		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.CRMAt{RefId: body.CRM_ID, CRMContent: body.CRMContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating CRMat")
	}
	key := services.GetKey(models.CRMView{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}

func UpdateCRM(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.CRM, int, error) {
	var body models.CRM
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("Body:", body)

	conditions = map[string]interface{}{
		"crm_id": body.CRM_ID,
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed updating CRM")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.CRMAt{RefId: body.CRM_ID, CRMContent: body.CRMContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating CRMat")
	}
	fmt.Println("body: ", body)
	key := services.GetKey(models.CRMView{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}
