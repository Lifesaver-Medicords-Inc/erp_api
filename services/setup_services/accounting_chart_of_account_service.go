package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
)

// "errors"

// func GetChartOfAccounts(conditions map[string]interface{}) ([]models.ChartOfAccount, int, error) {

// 	var based_service = services.NewInMemoryRepository(nil, nil, models.ChartOfAccount{}, models.ChartOfAccountAt{})

// 	return based_service.FetchAll()
// }
// func GetChartOfAccount(id int) (models.ChartOfAccount, int, error) {

// 	conditions := map[string]interface{}{
// 		"id": id,
// 	}

// 	var Group models.ChartOfAccount

// 	if err := services.DbGet(&Group, conditions); err != nil {
// 		return Group, fiber.StatusInternalServerError, errors.New("failed getting Group")
// 	}

// 	return Group, 0, nil
// }

// func CreateChartOfAccount(c *fiber.Ctx, tx *gorm.DB) (models.ChartOfAccount, int, error) {

// 	var based_service = services.NewInMemoryRepository(c, tx, models.ChartOfAccount{}, models.ChartOfAccountAt{})

// 	var body models.ChartOfAccount
// 	if err := c.BodyParser(&body); err != nil {

// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.ChartOfAccountAt{RefId: body.ID, AccountCode: body.AccountCode, At: at}

// 	return based_service.Create(body, atdata)
// }

// func UpdateChartOfAccount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ChartOfAccount, int, error) {

// 	var based_service = services.NewInMemoryRepository(c, tx, models.ChartOfAccount{}, models.ChartOfAccountAt{})

// 	var body models.ChartOfAccount
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.ChartOfAccountAt{RefId: body.ID, AccountCode: body.AccountCode, At: at}

// 	return based_service.Update(body, atdata, conditions)
// }

// func DeleteChartOfAccount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ChartOfAccount, int, error) {

// 	var body models.ChartOfAccount
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	if err := services.DbDelete(tx, &body, nil); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed deleting Group")
// 	}

// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	atdata := models.ChartOfAccountAt{RefId: body.ID, AccountCode: body.AccountCode, At: at}

// 	if err := services.DbInsert(tx, &atdata); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed creating Groupat")
// 	}

// 	return body, 0, nil
// }

func GetChartOfAccounts(conditions map[string]interface{}) ([]models.ChartOfAccountViewList, int, error) {
	var response []models.ChartOfAccountViewList

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed to get chart of account")
	}
	return response, 0, nil
}
