package sales_services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

type CustomerBody struct {
	models.Bpi
	General  models.BpiGeneral  `json:"general"`
	Contacts models.BpiContacts `json:"contacts"`
	Address  models.BpiAddress  `json:"address"`
}

type FinalizeBody struct {
	models.SalesQuotation
	// Child 1 - previously omitted, so any line-item edits sent alongside a
	// PUT /sales/quotation request were parsed away by BodyParser and
	// silently never persisted, even though the request appeared to
	// succeed. Line items are only synced if this is present in the body.
	SalesQuotationQuickWithImages []SalesQuotationQuickWithImages `json:"sales_quotation_quick"`
}

type Body struct {
	models.SalesQuotation
	// Child 1 - must be a slice: a quotation normally has multiple line
	// items, and GORM's Find into a single struct scans each matching row
	// into the same destination, silently keeping only the last one.
	SalesQuotationQuick []models.SalesQuotationQuick `json:"sales_quotation_quick"`
}

type CreateBody struct {
	models.SalesQuotation
	// Child 1
	SalesQuotationQuickWithImages []SalesQuotationQuickWithImages `json:"sales_quotation_quick"`
}

type SalesQuotationQuickWithImages struct {
	models.SalesQuotationQuick
	QuickSelectedImage []models.SalesQuotationSelectedImage `json:"quick_selected_image"`
}

type UpdateQuickSelectedImage struct {
	NewSelectedImage    []models.SalesQuotationSelectedImageContent `json:"new_selected_image"`
	UpdateSelectedImage []models.SalesQuotationSelectedImage        `json:"update_selected_image"`
}

type ItemBody struct {
	Items    models.Item `json:"items"`
	ItemName models.Name `json:"item_name"`
}

func GetLatestQuotations(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		LatestQuote         []models.LatestQuotationView
		SalesQuotationQuick []models.SalesQuotationQuick
	}
	var response Response
	if err := services.DbGet(&response.LatestQuote, conditions); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed getting latest quotations")
	}

	var filteredQuotations []models.LatestQuotationView
	for _, quotation := range response.LatestQuote {
		if quotation.ProjectName == "" {
			filteredQuotations = append(filteredQuotations, quotation)
		}
	}

	response.LatestQuote = filteredQuotations

	if err := GetSalesQuotationQuicks(&response.SalesQuotationQuick, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetSalesQuotations(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		SalesQuotation               []models.SalesQuotation
		SalesQuotationQuick          []models.SalesQuotationQuick
		SalesQuotationSelectedImages []models.SalesQuotationSelectedImage
	}

	var response Response

	if err := services.DbGet(&response.SalesQuotation, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting sales quotations")
	}

	var filteredQuotations []models.SalesQuotation
	for _, quotation := range response.SalesQuotation {
		if quotation.ProjectName == "" {
			filteredQuotations = append(filteredQuotations, quotation)
		}
	}

	response.SalesQuotation = filteredQuotations

	if err := GetSalesQuotationQuicks(&response.SalesQuotationQuick, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	if err := GetQuotationQuickSelectedImages(&response.SalesQuotationSelectedImages, conditions); err != nil {
		return response, fiber.StatusInternalServerError, err
	}

	return response, 0, nil
}

func GetSalesQuotation(id int) (Body, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var record Body

	if err := services.DbGet(&record.SalesQuotation, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	// DbGet uses GORM's Find, which does not error when nothing matches, so
	// a nonexistent id would otherwise fall through and report success with
	// an empty record. Check explicitly and report 404 instead.
	if record.SalesQuotation.ID == 0 {
		return record, fiber.StatusNotFound, errors.New("quotation not found")
	}

	conditions = map[string]interface{}{
		// based on parent ID
		"based_id": record.ID,
	}

	// Child 1
	if err := GetSalesQuotationQuicks(&record.SalesQuotationQuick, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

func CreateSalesQuotation(c *fiber.Ctx, tx *gorm.DB) (CreateBody, int, error) {
	var body CreateBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("pt 1", err)
		fmt.Println("ERR", body)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbInsert(tx, &body.SalesQuotation); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", body)
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	atdata := models.SalesQuotationAt{RefId: body.ID, SalesQuotationContent: body.SalesQuotationContent, At: at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	for _, v := range body.SalesQuotationQuickWithImages {
		if err := CreateSalesQuotationQuick(tx, body.ID, v.SalesQuotationQuick, v.QuickSelectedImage, at, body.ValidUntil); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}
func GetBpiCustomers(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		GetBpiCustomer []models.BpiCustomerView
	}

	var response Response

	if err := services.DbGet(&response.GetBpiCustomer, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi customer list")
	}

	return response, 0, nil
}

func GetBpis(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		GetBpiCustomer []models.BpiCustomerView
	}

	var response Response

	if err := services.DbGet(&response.GetBpiCustomer, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting bpi customer list")
	}

	return response, 0, nil
}

func GetBpi(id int) (CustomerBody, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}

	var record CustomerBody

	if err := services.DbGet(&record.Bpi, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	if record.Bpi.ID == 0 {
		return record, fiber.StatusNotFound, errors.New("bpi not found")
	}

	conditions = map[string]interface{}{
		"based_id": record.ID,
	}

	if err := services.DbGet(&record.General, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := services.DbGet(&record.Address, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := services.DbGet(&record.Contacts, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}
	fmt.Println("DATA", record)
	return record, 0, nil
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items           []models.Item            `json:"items"`
		ItemSpecs       []models.ItemSpecs       `json:"itemspecs"`
		AdditionalSpecs []models.AdditionalSpecs `json:"additionalspecs"`
	}

	var response Response

	if err := services.DbGet(&response.Items, conditions); err != nil {
		return response, fiber.StatusInternalServerError, errors.New("failed getting items")
	}

	fmt.Println(response)
	return response, 0, nil
}

func GetItem(id int) (ItemBody, int, error) {
	conditions := map[string]interface{}{
		"item_name_id": id,
	}

	var record ItemBody

	if err := services.DbGet(&record.Items, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}

	conditions = map[string]interface{}{
		"id": id,
	}

	if err := services.DbGet(&record.ItemName, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if record.ItemName.ID == 0 {
		return record, fiber.StatusNotFound, errors.New("item not found")
	}

	return record, 0, nil
}

func UpdateFinalizeQuote(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (FinalizeBody, int, error) {
	var body FinalizeBody

	fmt.Print(body)

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Guard: without a valid id, DbUpdate's implicit primary-key match falls
	// through to an unconditioned UPDATE that touches every row in the table.
	// Fail fast instead of silently overwriting the whole quotation list.
	if body.SalesQuotation.ID == 0 {
		return body, fiber.StatusBadRequest, errors.New("missing or invalid quotation id")
	}

	// Guard: a header-only quotation with nothing to price, produce, or
	// eventually invoice used to finalize successfully with zero line items.
	// Scoped to finalize specifically (IsFinalized true), not every plain
	// edit/save - a draft may legitimately be saved mid-entry before any
	// item has been picked yet (Sales Quotation, §5.1.2: a new quote opens
	// with one blank, unselected row).
	if body.SalesQuotation.IsFinalized && len(body.SalesQuotationQuickWithImages) == 0 {
		return body, fiber.StatusBadRequest, errors.New("cannot finalize a quotation with no line items")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	// Always scope the update to this specific id explicitly, rather than
	// relying solely on GORM's implicit primary-key condition.
	updateConditions := map[string]interface{}{"id": body.SalesQuotation.ID}
	for k, v := range conditions {
		updateConditions[k] = v
	}

	if err := UpdateQuotationQuick(tx, body.SalesQuotation, at, updateConditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// Sync any line items sent alongside the header: existing rows (id set)
	// get updated in place, new rows (id zero) get inserted.
	for _, v := range body.SalesQuotationQuickWithImages {
		quick := v.SalesQuotationQuick
		if quick.ID == 0 {
			if err := CreateSalesQuotationQuick(tx, body.SalesQuotation.ID, quick, v.QuickSelectedImage, at, body.SalesQuotation.ValidUntil); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
			continue
		}

		itemConditions := map[string]interface{}{"id": quick.ID}
		if err := UpdateSalesQuotationQuick(tx, quick, at, itemConditions, body.SalesQuotation.ValidUntil); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	return body, 0, nil
}

// for finalizing the quotation
func UpdateQuotationQuick(tx *gorm.DB, Quotation models.SalesQuotation, at models.At, conditions map[string]interface{}) error {
	if err := services.DbUpdate(tx, &Quotation, conditions); err != nil {
		return errors.New("failed updating quotation")
	}

	quotationat := models.SalesQuotationAt{
		RefId:                 Quotation.ID,
		SalesQuotationContent: Quotation.SalesQuotationContent,
		At:                    at,
	}

	if err := services.DbInsert(tx, &quotationat); err != nil {
		return errors.New("failed creating project content")
	}

	return nil
}
