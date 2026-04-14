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

// func UpdateSalesProject(c *fiber.Ctx, tx *gorm.DB) (CreateProjectBody, int, error) {
// 	var body CreateProjectBody
// 	if err := c.BodyParser(&body); err != nil {
// 		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
// 	}

// 	if body.ID == 0 {
// 		return body, fiber.StatusBadRequest, errors.New("invalid sales project id")
// 	}

// 	// ---- UPDATE SALES QUOTATION ----
// 	if err := services.DbUpdate(tx, &body.SalesQuotation, map[string]interface{}{
// 		"id": body.ID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed updating sales project")
// 	}

// 	// ---- AT CONTEXT ----
// 	at, ok := c.Locals("at").(models.At)
// 	if !ok {
// 		at = models.At{}
// 	}

// 	// ---- UPDATE SALES QUOTATION AT ----
// 	atData := models.SalesQuotationAt{
// 		RefId:                 body.ID,
// 		SalesQuotationContent: body.SalesQuotationContent,
// 		At:                    at,
// 	}

// 	if err := services.DbUpdate(tx, &atData, map[string]interface{}{
// 		"ref_id": body.ID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, errors.New("failed updating sales quotation at")
// 	}

// 	// ---- MULTIPLIERS (RESET) ----
// 	if err := services.DbDelete(tx, &models.SalesProjectMultiplier{}, map[string]interface{}{
// 		"based_id": body.ID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	for _, v := range body.SalesProjectMultiplier {
// 		if err := CreateSalesProjectMultiplier(tx, body.ID, v, at); err != nil {
// 			return body, fiber.StatusInternalServerError, err
// 		}
// 	}

// 	// ---- HISTORY (RESET) ----
// 	if err := services.DbDelete(tx, &models.SalesProjectHistory{}, map[string]interface{}{
// 		"based_id": body.ID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	for _, v := range body.SalesProjectHistory {
// 		if err := CreateSalesProjectHistory(tx, body.ID, v, at); err != nil {
// 			return body, fiber.StatusInternalServerError, err
// 		}
// 	}

// 	// ---- ITEM SET ----
// 	if err := UpdateProjectItemSet(tx, body.SalesProjectItemSet, at, map[string]interface{}{}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	itemSetID := body.SalesProjectItemSet.ItemSetID

// 	// ---- CONTENT ----
// 	if err := services.DbDelete(tx, &models.SalesProjectContent{}, map[string]interface{}{
// 		"item_set_id": itemSetID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	if err := CreateProjectContent(tx, itemSetID, body.SalesProjectContent, at); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	// ---- ADVANCED CONDITIONS ----
// 	if err := services.DbDelete(tx, &models.SalesProjectAdvancedConditions{}, map[string]interface{}{
// 		"item_set_id": itemSetID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	if err := CreateProjectAdvancedConditions(
// 		tx,
// 		itemSetID,
// 		body.SalesProjectContentAdvancedCondition,
// 		at,
// 	); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	// ---- ITEMS ----
// 	if err := services.DbDelete(tx, &models.SalesProjectItems{}, map[string]interface{}{
// 		"item_set_id": itemSetID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	for _, v := range body.SalesProjectItems {
// 		if err := CreateProjectItems(tx, itemSetID, v, at); err != nil {
// 			return body, fiber.StatusInternalServerError, err
// 		}
// 	}

// 	// ---- WIRINGS ----
// 	if err := services.DbDelete(tx, &models.SalesProjectWiring{}, map[string]interface{}{
// 		"item_set_id": itemSetID,
// 	}); err != nil {
// 		return body, fiber.StatusInternalServerError, err
// 	}

// 	for _, v := range body.SalesProjectWirings {
// 		if err := CreateProjectWiring(tx, itemSetID, v, at); err != nil {
// 			return body, fiber.StatusInternalServerError, err
// 		}
// 	}

// 	return body, fiber.StatusOK, nil
// }

func UpdateSalesProject(c *fiber.Ctx, tx *gorm.DB) (CreateProjectBody, int, error) {
	var body CreateProjectBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, fmt.Errorf("cannot bind request: %w", err)
	}

	if body.ID == 0 {
		return body, fiber.StatusBadRequest, errors.New("invalid sales project id")
	}

	// ---- UPDATE SALES QUOTATION ----
	if err := services.DbUpdate(tx, &body.SalesQuotation, map[string]interface{}{
		"id": body.ID,
	}); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("update sales quotation: %w", err)
	}

	// ---- AT CONTEXT ----
	at, _ := c.Locals("at").(models.At)

	// ---- UPDATE SALES QUOTATION AT ----
	atData := models.SalesQuotationAt{
		RefId:                 body.ID,
		SalesQuotationContent: body.SalesQuotationContent,
		At:                    at,
	}

	if err := services.DbUpdate(tx, &atData, map[string]interface{}{
		"ref_id": body.ID,
	}); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("update quotation at: %w", err)
	}

	// ---- SYNC MULTIPLIERS ----
	if err := SyncSalesProjectMultipliers(tx, body.ID, body.SalesProjectMultiplier, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- SYNC HISTORY ----
	if err := SyncSalesProjectHistory(tx, body.ID, body.SalesProjectHistory, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- ITEM SET ----
	body.SalesProjectItemSet.BasedId = body.ID
	if err := UpdateProjectItemSet(tx, body.SalesProjectItemSet, at, map[string]interface{}{
		"itemset_id": body.SalesProjectItemSet.ItemSetID,
	}); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	Id := body.ID

	// ---- SYNC CONTENT ----
	if err := SyncProjectContent(tx, Id, body.SalesProjectContent, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- SYNC ADVANCED CONDITIONS ----
	if err := SyncProjectAdvancedConditions(tx, Id, body.SalesProjectContentAdvancedCondition, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- SYNC ITEMS ----
	if err := SyncProjectItems(tx, Id, body.SalesProjectItems, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	fmt.Print("wiring")
	fmt.Println(Id)
	// ---- SYNC WIRINGS ----
	if err := SyncProjectWirings(tx, Id, body.SalesProjectWirings, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	return body, fiber.StatusOK, nil
}

func SyncSalesProjectMultipliers(
	tx *gorm.DB,
	projectID uint,
	incoming []models.SalesProjectMultiplier,
	at models.At,
) error {

	var existing []models.SalesProjectMultiplier
	if err := tx.
		Where("based_id = ?", projectID).
		Find(&existing).Error; err != nil {
		return err
	}

	// Map existing by MultiplierID
	existingMap := make(map[uint]models.SalesProjectMultiplier)
	for _, e := range existing {
		existingMap[e.MultiplierID] = e
	}

	keepIDs := make(map[uint]bool)

	for _, v := range incoming {
		v.BasedId = projectID

		if v.MultiplierID != 0 {
			// ---- UPDATE ----
			if _, ok := existingMap[v.MultiplierID]; ok {

				if err := tx.Model(&models.SalesProjectMultiplier{}).
					Where("multiplier_id = ?", v.MultiplierID).
					Updates(map[string]interface{}{
						"brand":       v.Brand,
						"component":   v.Component,
						"description": v.Description,
						"multiplier":  v.Multiplier,
					}).Error; err != nil {
					return err
				}

				// ---- UPSERT AT ----
				atData := models.SalesProjectMultiplierAt{
					RefId:                         v.MultiplierID,
					SalesProjectMultiplierContent: v.SalesProjectMultiplierContent,
					At:                            at,
				}

				if err := tx.
					Where("ref_id = ?", v.MultiplierID).
					Assign(atData).
					FirstOrCreate(&atData).Error; err != nil {
					return err
				}

				keepIDs[v.MultiplierID] = true
				continue
			}
		}

		// ---- CREATE ----
		v.MultiplierID = 0

		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectMultiplierAt{
			RefId:                         v.MultiplierID,
			SalesProjectMultiplierContent: v.SalesProjectMultiplierContent,
			At:                            at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		keepIDs[v.MultiplierID] = true
	}

	// ---- DELETE REMOVED ----
	for _, e := range existing {
		if !keepIDs[e.MultiplierID] {
			if err := tx.Delete(&models.SalesProjectMultiplier{}, e.MultiplierID).Error; err != nil {
				return err
			}

			// optional: also delete AT
			if err := tx.
				Where("ref_id = ?", e.MultiplierID).
				Delete(&models.SalesProjectMultiplierAt{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func SyncSalesProjectHistory(
	tx *gorm.DB,
	projectID uint,
	incoming []models.SalesProjectHistory,
	at models.At,
) error {

	var existing []models.SalesProjectHistory
	if err := tx.
		Where("based_id = ?", projectID).
		Find(&existing).Error; err != nil {
		return err
	}

	// Map existing by HistoryID
	existingMap := make(map[uint]models.SalesProjectHistory)
	for _, e := range existing {
		existingMap[e.HistoryID] = e
	}

	keepIDs := make(map[uint]bool)

	for _, v := range incoming {
		v.BasedId = projectID

		if v.HistoryID != 0 {
			// ---- UPDATE ----
			if _, ok := existingMap[v.HistoryID]; ok {

				if err := tx.Model(&models.SalesProjectHistory{}).
					Where("history_id = ?", v.HistoryID).
					Updates(map[string]interface{}{
						"user":     v.User,
						"date":     v.Date,
						"time":     v.Time,
						"old_data": v.OldData,
						"new_data": v.NewData,
					}).Error; err != nil {
					return err
				}

				// ---- UPSERT AT ----
				atData := models.SalesProjectHistoryAt{
					RefId: v.HistoryID,
					SalesProjectHistoryContent: models.SalesProjectHistoryContent{
						BasedId: projectID,
						User:    v.User,
						Date:    v.Date,
						Time:    v.Time,
						OldData: v.OldData,
						NewData: v.NewData,
					},
					At: at,
				}

				if err := tx.
					Where("ref_id = ?", v.HistoryID).
					Assign(atData).
					FirstOrCreate(&atData).Error; err != nil {
					return err
				}

				keepIDs[v.HistoryID] = true
				continue
			}
		}

		// ---- CREATE ----
		v.HistoryID = 0

		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectHistoryAt{
			RefId: v.HistoryID,
			SalesProjectHistoryContent: models.SalesProjectHistoryContent{
				BasedId: projectID,
				User:    v.User,
				Date:    v.Date,
				Time:    v.Time,
				OldData: v.OldData,
				NewData: v.NewData,
			},
			At: at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		keepIDs[v.HistoryID] = true
	}

	// ---- DELETE REMOVED ----
	for _, e := range existing {
		if !keepIDs[e.HistoryID] {
			if err := tx.
				Delete(&models.SalesProjectHistory{}, e.HistoryID).Error; err != nil {
				return err
			}

			// optional: also delete AT
			if err := tx.
				Where("ref_id = ?", e.HistoryID).
				Delete(&models.SalesProjectHistoryAt{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func SyncProjectContent(
	tx *gorm.DB,
	itemSetID uint,
	incoming models.SalesProjectContent,
	at models.At,
) error {

	var existing []models.SalesProjectContent
	if err := tx.
		Preload("SalesProjectContentFinal").
		Where("based_id = ?", itemSetID).
		Find(&existing).Error; err != nil {
		return err
	}

	// Map existing content
	existingMap := make(map[uint]models.SalesProjectContent)
	for _, e := range existing {
		existingMap[e.ContentID] = e
	}

	keepIDs := make(map[uint]bool)

	v := incoming
	v.BasedId = itemSetID

	if v.ContentID != 0 {
		// ---- UPDATE CONTENT ----
		if old, ok := existingMap[v.ContentID]; ok {

			if err := tx.Model(&models.SalesProjectContent{}).
				Where("content_id = ?", v.ContentID).
				Updates(map[string]interface{}{
					"item_designation":     v.ItemDesignation,
					"application":          v.Application,
					"additional":           v.Additional,
					"flow":                 v.Flow,
					"head":                 v.Head,
					"voltage":              v.Voltage,
					"rpm":                  v.RPM,
					"hp":                   v.HP,
					"phase":                v.Phase,
					"no_of_sets":           v.NoOfSets,
					"no_of_pump_set":       v.NoOfPumpSet,
					"item_set_description": v.ItemSetDescription,
					"item_set_notes":       v.ItemSetNotes,
					"template_project_id":  v.TemplateProjectId,
					"is_wiring":            v.IsWiring,
				}).Error; err != nil {
				return err
			}

			// ---- UPSERT AT ----
			atData := models.SalesProjectContentAt{
				RefID:                      v.ContentID,
				SalesProjectContentContent: v.SalesProjectContentContent,
				At:                         at,
			}

			if err := tx.
				Where("ref_id = ?", v.ContentID).
				Assign(atData).
				FirstOrCreate(&atData).Error; err != nil {
				return err
			}

			// ---- SYNC CHILDREN ----
			if err := SyncProjectContentFinals(
				tx,
				v.ContentID,
				old.SalesProjectContentFinal,
				v.SalesProjectContentFinal,
			); err != nil {
				return err
			}

			keepIDs[v.ContentID] = true
		}
	} else {
		// ---- CREATE CONTENT ----
		v.ContentID = 0
		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectContentAt{
			RefID:                      v.ContentID,
			SalesProjectContentContent: v.SalesProjectContentContent,
			At:                         at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		// ---- CREATE CHILDREN ----
		for i := range v.SalesProjectContentFinal {
			v.SalesProjectContentFinal[i].SalesProjectContentID = v.ContentID
		}

		if len(v.SalesProjectContentFinal) > 0 {
			if err := tx.Create(&v.SalesProjectContentFinal).Error; err != nil {
				return err
			}
		}

		keepIDs[v.ContentID] = true
	}

	// ---- DELETE REMOVED CONTENT ----
	for _, e := range existing {
		if !keepIDs[e.ContentID] {

			// delete children first (FK safety)
			if err := tx.
				Where("sales_project_content_id = ?", e.ContentID).
				Delete(&models.SalesProjectContentFinal{}).Error; err != nil {
				return err
			}

			// delete AT records
			if err := tx.
				Where("ref_id = ?", e.ContentID).
				Delete(&models.SalesProjectContentAt{}).Error; err != nil {
				return err
			}

			if err := tx.
				Delete(&models.SalesProjectContent{}, e.ContentID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func SyncProjectContentFinals(
	tx *gorm.DB,
	contentID uint,
	existing []models.SalesProjectContentFinal,
	incoming []models.SalesProjectContentFinal,
) error {

	// Map existing by ID
	existingMap := make(map[uint]models.SalesProjectContentFinal)
	for _, e := range existing {
		existingMap[e.ID] = e
	}

	keepIDs := make(map[uint]bool)

	for _, v := range incoming {
		v.SalesProjectContentID = contentID

		if v.ID != 0 {
			// ---- UPDATE ----
			if _, ok := existingMap[v.ID]; ok {

				if err := tx.Model(&models.SalesProjectContentFinal{}).
					Where("id = ?", v.ID).
					Updates(map[string]interface{}{
						"final":   v.Final,
						"fla":     v.Fla,
						"voltage": v.Voltage,
					}).Error; err != nil {
					return err
				}

				keepIDs[v.ID] = true
				continue
			}
		}

		// ---- CREATE ----
		v.ID = 0

		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		keepIDs[v.ID] = true
	}

	// ---- DELETE REMOVED ----
	for _, e := range existing {
		if !keepIDs[e.ID] {
			if err := tx.Delete(&models.SalesProjectContentFinal{}, e.ID).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func SyncProjectAdvancedConditions(
	tx *gorm.DB,
	itemSetID uint,
	incoming models.SalesProjectAdvancedConditions,
	at models.At,
) error {

	var existing models.SalesProjectAdvancedConditions

	// Try to find existing record
	found := tx.
		Where("based_id = ?", itemSetID).
		First(&existing).Error

	if found != nil && found != gorm.ErrRecordNotFound {
		return found
	}

	// If no existing record, create it
	if found == gorm.ErrRecordNotFound {
		incoming.ConditionsID = 0
		incoming.BasedId = itemSetID

		if err := tx.Create(&incoming).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectAdvancedConditionsAt{
			RefID:                                 incoming.ConditionsID,
			SalesProjectAdvancedConditionsContent: incoming.SalesProjectAdvancedConditionsContent,
			At:                                    at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		return nil
	}

	// ---- UPDATE ----
	if err := tx.Model(&models.SalesProjectAdvancedConditions{}).
		Where("conditions_id = ?", existing.ConditionsID).
		Updates(map[string]interface{}{
			"pump_brand":              incoming.PumpBrand,
			"driver_type":             incoming.DriverType,
			"pressure":                incoming.Pressure,
			"motor_enclosure":         incoming.MotorEnclosure,
			"motor_manufacturer":      incoming.MotorManufacturer,
			"liquid_type":             incoming.LiquidType,
			"controller_manufacturer": incoming.ControllerManufacturer,
			"starting_method":         incoming.StartingMethod,
			"suction_size":            incoming.SuctionSize,
			"discharge_size":          incoming.DischargeSize,
		}).Error; err != nil {
		return err
	}

	// ---- UPSERT AT ----
	atData := models.SalesProjectAdvancedConditionsAt{
		RefID:                                 existing.ConditionsID,
		SalesProjectAdvancedConditionsContent: incoming.SalesProjectAdvancedConditionsContent,
		At:                                    at,
	}

	if err := tx.
		Where("ref_id = ?", existing.ConditionsID).
		Assign(atData).
		FirstOrCreate(&atData).Error; err != nil {
		return err
	}

	return nil
}

func SyncProjectItems(
	tx *gorm.DB,
	itemSetID uint,
	incoming []models.SalesProjectItems,
	at models.At,
) error {

	var existing []models.SalesProjectItems
	if err := tx.
		Where("based_id = ?", itemSetID).
		Find(&existing).Error; err != nil {
		return err
	}

	// Map existing by ItemsID
	existingMap := make(map[uint]models.SalesProjectItems)
	for _, e := range existing {
		existingMap[e.ItemsID] = e
	}

	keepIDs := make(map[uint]bool)

	for _, v := range incoming {
		v.BasedId = itemSetID

		if v.ItemsID != 0 {
			// ---- UPDATE ----
			if _, ok := existingMap[v.ItemsID]; ok {

				if err := tx.Model(&models.SalesProjectItems{}).
					Where("items_id = ?", v.ItemsID).
					Updates(map[string]interface{}{
						"template_id":         v.TemplateID,
						"bom_id":              v.BomID,
						"item_id":             v.ItemID,
						"reference_code":      v.ReferenceCode,
						"man_days":            v.ManDays,
						"labor_rate":          v.LaborRate,
						"components":          v.Components,
						"model":               v.Model,
						"item_inv_type":       v.ItemInvType,
						"qty":                 v.Qty,
						"multiplier":          v.Multiplier,
						"discount_price":      v.DiscountPrice,
						"list_price_per_unit": v.ListPricePerUnit,
						"component_total":     v.ComponentTotal,
						"notes":               v.Notes,
					}).Error; err != nil {
					return err
				}

				// ---- UPSERT AT ----
				atData := models.SalesProjectItemsAt{
					RefID:                    v.ItemsID,
					SalesProjectItemsContent: v.SalesProjectItemsContent,
					At:                       at,
				}

				if err := tx.
					Where("ref_id = ?", v.ItemsID).
					Assign(atData).
					FirstOrCreate(&atData).Error; err != nil {
					return err
				}

				keepIDs[v.ItemsID] = true
				continue
			}
		}

		// ---- CREATE ----
		v.ItemsID = 0

		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectItemsAt{
			RefID:                    v.ItemsID,
			SalesProjectItemsContent: v.SalesProjectItemsContent,
			At:                       at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		keepIDs[v.ItemsID] = true
	}

	// ---- DELETE REMOVED ----
	for _, e := range existing {
		if !keepIDs[e.ItemsID] {
			if err := tx.Delete(&models.SalesProjectItems{}, e.ItemsID).Error; err != nil {
				return err
			}

			// optional: also delete AT
			if err := tx.
				Where("ref_id = ?", e.ItemsID).
				Delete(&models.SalesProjectItemsAt{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

func SyncProjectWirings(
	tx *gorm.DB,
	itemSetID uint,
	incoming []models.SalesProjectWiring,
	at models.At,
) error {

	var existing []models.SalesProjectWiring
	if err := tx.
		Where("based_id = ?", itemSetID).
		Find(&existing).Error; err != nil {
		return err
	}

	// Map existing by ID
	existingMap := make(map[uint]models.SalesProjectWiring)
	for _, e := range existing {
		existingMap[e.ID] = e
	}

	keepIDs := make(map[uint]bool)

	for _, v := range incoming {
		v.BasedId = itemSetID

		if v.ID != 0 {
			// ---- UPDATE ----
			if _, ok := existingMap[v.ID]; ok {

				if err := tx.Model(&models.SalesProjectWiring{}).
					Where("id = ?", v.ID).
					Updates(map[string]interface{}{
						"materials":              v.Materials,
						"amp_req":                v.AmpReq,
						"wire_req":               v.WireReq,
						"description":            v.Description,
						"num_of_wires_set":       v.NumOfWiresSet,
						"num_of_qty_set":         v.NumOfQtySet,
						"distance_travelled_set": v.DistanceTravelledSet,
						"allowance_wire_set":     v.AllowanceWireSet,
						"qty":                    v.Qty,
						"num_of_sets":            v.NumOfSets,
						"total_qty":              v.TotalQty,
						"cost":                   v.Cost,
						"total_cost":             v.TotalCost,
					}).Error; err != nil {
					return err
				}

				// ---- UPSERT AT ----
				atData := models.SalesProjectWiringAt{
					RefId:                     v.ID,
					SalesProjectWiringContent: v.SalesProjectWiringContent,
					At:                        at,
				}

				if err := tx.
					Where("ref_id = ?", v.ID).
					Assign(atData).
					FirstOrCreate(&atData).Error; err != nil {
					return err
				}

				keepIDs[v.ID] = true
				continue
			}
		}

		// ---- CREATE ----
		v.ID = 0

		if err := tx.Create(&v).Error; err != nil {
			return err
		}

		// ---- CREATE AT ----
		atData := models.SalesProjectWiringAt{
			RefId:                     v.ID,
			SalesProjectWiringContent: v.SalesProjectWiringContent,
			At:                        at,
		}

		if err := tx.Create(&atData).Error; err != nil {
			return err
		}

		keepIDs[v.ID] = true
	}

	// ---- DELETE REMOVED ----
	for _, e := range existing {
		if !keepIDs[e.ID] {
			if err := tx.Delete(&models.SalesProjectWiring{}, e.ID).Error; err != nil {
				return err
			}

			// optional: also delete AT
			if err := tx.
				Where("ref_id = ?", e.ID).
				Delete(&models.SalesProjectWiringAt{}).Error; err != nil {
				return err
			}
		}
	}

	return nil
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
