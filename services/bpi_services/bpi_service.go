package bpi_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type Body struct {
	models.Bpi
	IndustriesId   []uint                    `json:"industries_id"`
	General        models.BpiGeneralSchema   `json:"general"`
	Contacts       []models.BpiContacts      `json:"contacts"`
	Address        []models.BpiAddress       `json:"address"`
	Items          []models.BpiItems         `json:"items"`
	Finance        models.BpiFinance         `json:"finance"`
	Accreditations []models.BpiAccreditation `json:"accreditations"`
}

type TestBody struct {
	models.Bpi
	IndustriesId []uint                  `json:"industries_id"`
	General      models.BpiGeneralSchema `json:"general"`
	Contacts     []models.BpiContacts    `json:"contacts"`
	//	Address      []models.BpiAddress     `json:"address"`
	// Items        []models.BpiItems       `json:"items"`
	//Finance models.BpiFinance `json:"finance"`
	// Accreditations []models.BpiAccreditation `json:"accreditations"`
}

func GetBpis(conditions map[string]interface{}) (interface{}, int, error) {

	type Response struct {
		Bpi            []models.BpiView               `json:"bpi"`
		General        []models.BpiGeneralView        `json:"general"`
		Contacts       []models.BpiContactView        `json:"contacts"`
		Address        []models.BpiAddressView        `json:"address"`
		Items          []models.BpiItemsView          `json:"items"`
		Finance        []models.BpiFinance            `json:"finance"`
		FinancePending []models.BpiFinancePendingView `json:"finance_pending"`
		Accreditations []models.BpiAccreditationView  `json:"accreditations"`
		History        []models.BpiHistoryView        `json:"history"`
	}

	var response Response

	if err := services.DbGet(&response.Bpi, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpis")
	}

	if err := services.DbGet(&response.General, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi general list")
	}

	if err := services.DbGet(&response.Contacts, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi contacts")
	}

	if err := services.DbGet(&response.Address, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi address")
	}
	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi items")
	}
	if err := services.DbGet(&response.Finance, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting finance ")
	}
	if err := services.DbGet(&response.FinancePending, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting finance pending quotation")
	}
	if err := services.DbGet(&response.Accreditations, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting accreditations  ")
	}
	if err := services.DbGet(&response.History, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting history")
	}

	return response, 0, nil
}

func GetBpiUsers(employeeId string) (interface{}, int, error) {

	conditions := map[string]interface{}{
		"EmployeeId": employeeId,
	}
	var response []models.User

	if err := services.DbRaw(&response, "sp_GetEmployeeByType", conditions); err != nil {

		return response, fiber.StatusInternalServerError, errors.New("failed getting user data")
	}

	return response, 0, nil

}

func GetBpiEntityRecords(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.BpiEntityCount

	if err := services.DbGet(&response, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi entity counts")
	}
	return response, 0, nil
}

func GetBpiItemList(conditions map[string]interface{}) (interface{}, int, error) {
	var response []models.BpiItemList

	if err := services.DbGet(&response, conditions); err != nil {

		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi item list")
	}

	return response, 0, nil
}

func CreateBpi(c *fiber.Ctx, tx *gorm.DB) (Body, int, error) {
	var body Body
	var count int64

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := tx.Model(&body.Bpi).Where("id =?", body.ID).Count(&count).Error; err != nil {
		return body, fiber.StatusBadRequest, errors.New("failed to find data in Bpi Table")

	}

	at, ok := c.Locals("at").(models.At)
	userAt := utils.GetAtData(c, models.At{})
	at.AtUserId = userAt.AtUserId
	fmt.Println("atBPI", at)
	if !ok {
		at = models.At{}
	}

	// if record is not existing insert
	if count == 0 {

		if err := services.DbInsert(tx, &body.Bpi); err != nil {
			return body, fiber.StatusInternalServerError, errors.New(("failed creating parent"))
		}

		for _, v := range body.IndustriesId {
			if err := CreateBpiIndustries(tx, body.ID, uint(v), at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}

		parentat := models.BpiAt{RefId: body.ID, BpiContent: models.BpiContent{SalesId: body.SalesId, Name: body.Name, Tin: body.Tin, MainTelNo: body.MainTelNo}, At: at}
		if err := services.DbInsert(tx, &parentat); err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed creating parentat")
		}

		key := services.GetKey(models.BpiView{}, nil)
		services.InvalidateCache(key)
	}
	// invalidate all child keys
	InvalidateChildKey()

	//Create Bpi General

	if err := BpiGeneral(tx, body.ID, &body.General, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Create  Bpi Contacts

	for i := range body.Contacts {
		if err := BpiContact(tx, body.ID, body.General.ID, &body.Contacts[i], at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Create Bpi Address

	for _, v := range body.Address {

		if err := BpiAddress(tx, body.ID, body.General.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Create Bpi Items
	if err := CreateBpiItems(tx, body.ID, body.General.ID, body.Items, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	//Create Bpi Finance
	if err := CreateBpiFinance(tx, body.ID, body.General.ID, body.Finance, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}
	//Create Bpi Accreditations

	for _, v := range body.Accreditations {

		if err := CreateBpiAccreditation(tx, body.ID, body.General.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}
func UpdateBpi(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {

	var body Body

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	fmt.Println("BODY General REQUEST", body.General)
	fmt.Println("BODY Contacts  REQUEST", body.Contacts)
	fmt.Println("BODY  Address  REQUEST", body.Address)
	fmt.Println("BODY Items  REQUEST", body.Items)
	fmt.Println("BODY Items  REQUEST", body.Finance)

	if err := services.DbUpdate(tx, &body.Bpi, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New(("failed updating parent"))
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	parentat := models.BpiAt{RefId: body.ID, BpiContent: models.BpiContent{SalesId: body.SalesId, Name: body.Name, Tin: body.Tin, MainTelNo: body.MainTelNo}, At: at}
	if err := services.DbInsert(tx, &parentat); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating bpi at")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	IndustriesCondition := map[string]interface{}{
		"bpi_id": body.ID,
	}

	// Delete the bpi industries data  and Create data  in table
	if len(body.IndustriesId) != 0 {
		if err := services.DbDelete(tx, &models.BpiIndustries{}, IndustriesCondition); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	for _, v := range body.IndustriesId {
		if err := CreateBpiIndustries(tx, body.ID, uint(v), at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Update Bpi General

	if body.General.ID == 0 {

		if err := BpiGeneral(tx, body.ID, &body.General, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

	} else {
		err := UpdateBpiGeneral(tx, &body.General, at, conditions)
		if err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// Update  Bpi Contacts

	for _, v := range body.Contacts {

		if err := UpdateBpiContact(tx, body.General.ID, v, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

	}

	//Update Bpi Address

	for _, v := range body.Address {
		if err := UpdateBpiAddress(tx, body.General.ID, v, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	//Update Bpi Items

	if err := UpdateBpiItems(tx, body.ID, body.General.ID, body.Items, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Update Bpi Finance

	if err := UpdateBpiFinance(tx, body.Finance, at, body.ID); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	//Add  Bpi Accreditations not update

	for _, v := range body.Accreditations {
		fmt.Println("Body Accreditations", v)

		if err := CreateBpiAccreditation(tx, body.ID, body.General.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

	}

	return body, 0, nil
}
func InvalidateChildKey() {

	generalKey := services.GetKey(models.BpiGeneralView{}, nil)
	services.InvalidateCache(generalKey)

	contactKey := services.GetKey(models.BpiContactView{}, nil)
	services.InvalidateCache(contactKey)

	addressKey := services.GetKey(models.BpiAddressView{}, nil)
	services.InvalidateCache(addressKey)

	itemKey := services.GetKey(models.BpiItemsView{}, nil)
	services.InvalidateCache(itemKey)

	financeKey := services.GetKey(models.BpiFinance{}, nil)
	services.InvalidateCache(financeKey)

	financePendingKey := services.GetKey(models.BpiFinancePendingView{}, nil)
	services.InvalidateCache(financePendingKey)

	accreditationKey := services.GetKey(models.BpiAccreditationView{}, nil)
	services.InvalidateCache(accreditationKey)
}
