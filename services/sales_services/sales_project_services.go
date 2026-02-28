package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type ProjectBody struct {
	models.SalesQuotation
	SalesProjectMultiplier               models.SalesProjectMultiplier         `json:"sales_project_multiplier"`
	SalesProjectHistory                  models.SalesProjectHistory            `json:"sales_project_history"`
	SalesProjectItemSet                  models.SalesProjectItemSet            `json:"sales_project_item_set"`
	SalesProjectContent                  models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    models.SalesProjectItems              `json:"sales_project_items"`
	SalesProjectWirings                  models.SalesProjectWiring             `json:"sales_project_wiring"`
}

type AdvancedConditionRequest struct {
	Branch                               string                                `json:"branch"`
	ProjectId                            string                                `json:"project_id"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
}

type CreateProjectBody struct {
	models.SalesQuotation
	SalesProjectMultiplier               []models.SalesProjectMultiplier       `json:"sales_project_multiplier"`
	SalesProjectHistory                  []models.SalesProjectHistory          `json:"sales_project_history"`
	SalesProjectItemSet                  models.SalesProjectItemSet            `json:"sales_project_item_set"`
	SalesProjectContent                  models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    []models.SalesProjectItems            `json:"sales_project_items"`
	SalesProjectWirings                  []models.SalesProjectWiring           `json:"sales_project_wiring"`
}

type CreateNewProjectItem struct {
	SalesProjectItems models.SalesProjectItems `json:"sales_project_items"`
}

type CreateNewProjectItemz struct {
	SalesProjectItems []models.SalesProjectItems `json:"sales_project_items"`
}

type CreateNewProjectWiringBody struct {
	SalesProjectWirings []models.SalesProjectWiring `json:"sales_project_wiring"`
}

type CreateProjectBody2 struct {
	models.SalesQuotation
	SalesProjectMultiplier               []models.SalesProjectMultiplier       `json:"sales_project_multiplier"`
	SalesProjectHistory                  []models.SalesProjectHistory          `json:"sales_project_history"`
	SalesProjectItemSet                  []models.SalesProjectItemSet          `json:"sales_project_item_set"`
	SalesProjectContent                  models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    []models.SalesProjectItems            `json:"sales_project_items"`
	SalesProjectWirings                  models.SalesProjectWiring             `json:"sales_project_wiring"`
}

type UpdateProjectBody struct {
	models.SalesQuotation
	SalesProjectMultiplier               models.SalesProjectMultiplier         `json:"sales_project_multiplier"`
	SalesProjectHistory                  models.SalesProjectHistory            `json:"sales_project_history"`
	SalesProjectItemSet                  models.SalesProjectItemSet            `json:"sales_project_item_set"`
	SalesProjectContent                  models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    []models.SalesProjectItems            `json:"sales_project_items"`
	SalesProjectWirings                  []models.SalesProjectWiring           `json:"sales_project_wiring"`
}

func GetBpiSuppliers(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		BpiSuppliers []models.BpiSuppliersView
	}

	var response Response

	if err := services.DbGet(&response.BpiSuppliers, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi suppliers")
	}

	return response, 0, nil
}

func GetSalesProjectsWS(conditions map[string]interface{}, multiplierConditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		SalesQuotation                       []models.SalesQuotation
		SalesProjectMultiplier               []models.SalesProjectMultiplier         `json:"sales_project_multiplier"`
		SalesProjectHistory                  []models.SalesProjectHistory            `json:"sales_project_history"`
		SalesProjectItemSet                  []models.SalesProjectItemSet            `json:"sales_project_item_set"`
		SalesProjectContent                  []models.SalesProjectContent            `json:"sales_project_content"`
		SalesProjectContentAdvancedCondition []models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
		SalesProjectItems                    []models.SalesProjectItems              `json:"sales_project_items"`
		SalesProjectWirings                  []models.SalesProjectWiring             `json:"sales_project_wiring"`
	}

	var response Response

	if err := services.DbGetNoCache(&response.SalesQuotation, conditions); err != nil {
		fmt.Println("Error fetching SalesQuotation:", err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales projects")
	}

	var filteredQuotations []models.SalesQuotation
	for _, quotation := range response.SalesQuotation {
		if quotation.ProjectName != "" {
			filteredQuotations = append(filteredQuotations, quotation)
		}
	}

	response.SalesQuotation = filteredQuotations

	if err := services.DbGetNoCache(&response.SalesProjectMultiplier, multiplierConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := services.DbGetNoCache(&response.SalesProjectHistory, nil); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := services.DbGetNoCache(&response.SalesProjectItemSet, multiplierConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if len(response.SalesProjectItemSet) == 0 { 
		return response, fiber.StatusNotFound, errors.New("no item set found")
	}
	basedID := response.SalesProjectItemSet[0].ItemSetID

	itemSetChildConditions := map[string]interface{}{
		"based_id": basedID,
	}

	if err := services.DbGetNoCache(&response.SalesProjectContent, itemSetChildConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := services.DbGetNoCache(&response.SalesProjectContentAdvancedCondition, itemSetChildConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := services.DbGetNoCache(&response.SalesProjectItems, itemSetChildConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := services.DbGetNoCache(&response.SalesProjectWirings, itemSetChildConditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetSalesProjects(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		SalesQuotation                       []models.SalesQuotation
		SalesProjectMultiplier               []models.SalesProjectMultiplier         `json:"sales_project_multiplier"`
		SalesProjectHistory                  []models.SalesProjectHistory            `json:"sales_project_history"`
		SalesProjectItemSet                  []models.SalesProjectItemSet            `json:"sales_project_item_set"`
		SalesProjectContent                  []models.SalesProjectContent            `json:"sales_project_content"`
		SalesProjectContentAdvancedCondition []models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
		SalesProjectItems                    []models.SalesProjectItems              `json:"sales_project_items"`
		SalesProjectWirings                  []models.SalesProjectWiring             `json:"sales_project_wiring"`
	}

	var response Response

	if err := services.DbGet(&response.SalesQuotation, conditions); err != nil {
		fmt.Println("Error fetching SalesQuotation:", err)
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales projects")
	}

	var filteredQuotations []models.SalesQuotation
	for _, quotation := range response.SalesQuotation {
		if quotation.ProjectName != "" {
			filteredQuotations = append(filteredQuotations, quotation)
		}
	}

	response.SalesQuotation = filteredQuotations

	if err := GetSalesProjectMultiplier(&response.SalesProjectMultiplier, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetSalesProjectHistory(&response.SalesProjectHistory, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetProjectItemSet(&response.SalesProjectItemSet, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetSalesProjectContent(&response.SalesProjectContent, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetProjectAdvancedConditions(&response.SalesProjectContentAdvancedCondition, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := GetProjectItems(&response.SalesProjectItems, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}
	if err := GetProjectWiring(&response.SalesProjectWirings, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil

}

func CreateSalesProject(c *fiber.Ctx, tx *gorm.DB) (CreateProjectBody, int, error) {
	var body CreateProjectBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Print(err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesQuotation); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating projects")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	for _, v := range body.SalesProjectMultiplier {
		if err := CreateSalesProjectMultiplier(tx, body.ID, v, at); err != nil {
			fmt.Print("KEY ADVCOND SET::", body)
			return body, fiber.StatusInternalServerError, err
		}
	}

	for _, v := range body.SalesProjectHistory {
		if err := CreateSalesProjectHistory(tx, body.ID, v, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	if err := CreateProjectItemSet(tx, body.ID, &body.SalesProjectItemSet, at); err != nil {
		fmt.Print("KEY ITEM SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateProjectContent(tx, body.SalesProjectItemSet.ItemSetID, body.SalesProjectContent, at); err != nil {
		fmt.Print("KEY CONTENT SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateProjectAdvancedConditions(tx, body.SalesProjectItemSet.ItemSetID, body.SalesProjectContentAdvancedCondition, at); err != nil {
		fmt.Print("KEY ADVCOND SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	for _, v := range body.SalesProjectItems {
		if err := CreateProjectItems(tx, body.SalesProjectItemSet.ItemSetID, v, at); err != nil {

			return body, fiber.StatusInternalServerError, err
		}
	}

	for _, v := range body.SalesProjectWirings {
		if err := CreateProjectWiring(tx, body.SalesProjectItemSet.ItemSetID, v, at); err != nil {

			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func CreateNewProjectItemss(c *fiber.Ctx, tx *gorm.DB) (CreateNewProjectItemz, int, error) {
	var body CreateNewProjectItemz

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, item := range body.SalesProjectItems {
		if err := CreateProjectItems(tx, item.BasedId, item, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func CreateNewProjectWiring(c *fiber.Ctx, tx *gorm.DB) (CreateNewProjectWiringBody, int, error) {
	var body CreateNewProjectWiringBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, item := range body.SalesProjectWirings {
		if err := CreateProjectWiring(tx, item.BasedId, item, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

// CREATE NEW TAB WITH ITS CHILD ITEMS
func CreateNewItems(c *fiber.Ctx, tx *gorm.DB) (CreateProjectBody, int, error) {
	var body CreateProjectBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := CreateProjectItemSet(tx, body.SalesProjectItemSet.BasedId, &body.SalesProjectItemSet, at); err != nil {
		fmt.Print("KEY ITEM SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateProjectContent(tx, body.SalesProjectItemSet.ItemSetID, body.SalesProjectContent, at); err != nil {
		fmt.Print("KEY CONTENT SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	if err := CreateProjectAdvancedConditions(tx, body.SalesProjectItemSet.ItemSetID, body.SalesProjectContentAdvancedCondition, at); err != nil {
		fmt.Print("KEY ADVCOND SET::", err)
		return body, fiber.StatusInternalServerError, err
	}

	for _, v := range body.SalesProjectItems {
		if err := CreateProjectItems(tx, body.SalesProjectItemSet.ItemSetID, v, at); err != nil {

			return body, fiber.StatusInternalServerError, err
		}
	}

	for _, v := range body.SalesProjectWirings {
		if err := CreateProjectWiring(tx, body.SalesProjectItemSet.ItemSetID, v, at); err != nil {

			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func UpdateProjectMultiplier(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (UpdateProjectBody, int, error) {
	var body UpdateProjectBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	//services.BroadcastToProject()

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := UpdateSalesProjectMultiplier(tx, body.SalesProjectMultiplier, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateProjectContents(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (UpdateProjectBody, int, error) {
	var body UpdateProjectBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := UpdateProjectContent(tx, body.SalesProjectContent, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateProjectItemss(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (CreateNewProjectItemz, int, error) {
	var body CreateNewProjectItemz
	if err := c.BodyParser(&body); err != nil {
		fmt.Println(err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, item := range body.SalesProjectItems {
		if err := UpdateProjectItems(tx, item, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func UpdateProjectWirings(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (CreateNewProjectWiringBody, int, error) {
	var body CreateNewProjectWiringBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, item := range body.SalesProjectWirings {
		if err := UpdateProjectWiring(tx, item, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func UpdateProjectAdvancedCondition(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (AdvancedConditionRequest, int, error) {
	var body AdvancedConditionRequest
	if err := c.BodyParser(&body); err != nil {
		fmt.Println(err)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	fmt.Println("PARSED ADC::", body)
	if err := UpdateProjectAdvancedConditions(tx, &body.SalesProjectContentAdvancedCondition, at, conditions); err != nil {
		fmt.Println(err)
		return body, fiber.StatusInternalServerError, err
	}

	return body, 0, nil
}

func UpdateProjectItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (UpdateProjectBody, int, error) {
	var body UpdateProjectBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	for _, v := range body.SalesProjectItems {
		if err := UpdateProjectItems(tx, v, at, conditions); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

func GetItemPumps(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		ItemPumpsView []models.ItemPumpSpecsView
	}

	var response Response

	if err := services.DbGet(&response.ItemPumpsView, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting item pump view")
	}

	return response, 0, nil
}
