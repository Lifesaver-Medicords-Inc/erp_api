package setup_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
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
func CreateChartOfAccounts(c *fiber.Ctx, tx *gorm.DB) (models.ChartOfAccounts, int, error) {

	var body models.ChartOfAccounts
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating chart of accounts")
		}
		return body, fiber.StatusInternalServerError, err
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := models.ChartOfAccountsAt{RefId: body.ID, ChartOfAccountContent: models.ChartOfAccountContent{Code: body.Code, Name: body.Name, ClassId: body.ClassId, GroupId: body.GroupId}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	key := services.GetKey(models.ChartOfAccountViewList{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}

func UpdateChartOfAccount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ChartOfAccounts, int, error) {

	var body models.ChartOfAccounts
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbUpdate(tx, &body, conditions); err != nil {

		return body, fiber.StatusInternalServerError, err
	}
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := models.ChartOfAccountsAt{RefId: body.ID, ChartOfAccountContent: models.ChartOfAccountContent{Code: body.Code, Name: body.Name, ClassId: body.ClassId, GroupId: body.GroupId}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	key := services.GetKey(models.ChartOfAccountViewList{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}

func DeleteChartOfAccount(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.ChartOfAccounts, int, error) {

	var body models.ChartOfAccounts
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body, conditions); err != nil {

		return body, fiber.StatusInternalServerError, err
	}
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	atdata := models.ChartOfAccountsAt{RefId: body.ID, ChartOfAccountContent: models.ChartOfAccountContent{Code: body.Code, Name: body.Name, ClassId: body.ClassId, GroupId: body.GroupId}, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	key := services.GetKey(models.ChartOfAccountViewList{}, nil)
	services.InvalidateCache(key)
	return body, 0, nil
}
