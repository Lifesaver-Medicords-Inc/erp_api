package sales_services

import (
	"errors"
	"fmt"
	"strings"

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
	SalesProjectMultiplier []models.SalesProjectMultiplier `json:"sales_project_multiplier"`
	SalesProjectAllTabs    []SalesProjectAllTabs           `json:"sales_project_all_tabs"`
}
type SalesProjectAllTabs struct {
	SalesProjectHistory                  []models.SalesProjectHistory          `json:"sales_project_history"`
	SalesProjectItemSet                  models.SalesProjectItemSet            `json:"sales_project_item_set"`
	SalesProjectContent                  models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    []SalesProjectItemsWithImages         `json:"sales_project_items"`
	SalesProjectWirings                  []models.SalesProjectWiring           `json:"sales_project_wiring"`
}

// Project items didn't have any way to carry selected images through Create - this mirrors
// SalesQuotationQuickWithImages (the same pattern Quick Quote items already use): the item's
// own fields stay flattened at the top level (anonymous embed, no json tag), plus the
// per-item image selections riding alongside under "quick_selected_image".
type SalesProjectItemsWithImages struct {
	models.SalesProjectItems
	QuickSelectedImage []models.SalesQuotationSelectedImage `json:"quick_selected_image"`
}

type CreateNewProjectItem struct {
	SalesProjectItems models.SalesProjectItems `json:"sales_project_items"`
}

type NewProjectItem struct {
	SalesProjectItems []models.SalesProjectItems `json:"sales_project_items"`
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
	ID     uint       `json:"id"`
	Header HeaderDiff `json:"Header"`
	Tabs   []TabDiff  `json:"Tabs"`
}

type HeaderDiff struct {
	QuotationFields QuotationFieldChanges                         `json:"QuotationFields"`
	Multipliers     CollectionDiff[models.SalesProjectMultiplier] `json:"Multipliers"`
	// Change History entries generated client-side for edits that aren't scoped to any
	// one tab (top-part header fields like project name, and the multipliers grid) -
	// applied the same way as a tab's SalesProjectHistory, just keyed to the quotation's
	// own id (body.ID) instead of a tab's item_set_id.
	SalesProjectHistory CollectionDiff[models.SalesProjectHistory] `json:"SalesProjectHistory"`
}

type TabDiff struct {
	// The client already resolves and sends this: the tab's real item_set_id for an
	// existing tab, or a client-side placeholder (always > 0) for a brand-new tab that
	// hasn't been inserted yet. It was previously not read at all here, so the server fell
	// back to using the quotation's own id for every child record's based_id, which is
	// wrong per SalesProjectContentContent.BasedId's own comment ("SHOULD BE THE TAB # /
	// SET #") - see the resolution logic in UpdateSalesProject.
	BasedId                              uint                                                   `json:"BasedId"`
	SalesProjectHistory                  CollectionDiff[models.SalesProjectHistory]            `json:"SalesProjectHistory"`
	SalesProjectItemSet                  CollectionDiff[models.SalesProjectItemSet]            `json:"SalesProjectItemSet"`
	SalesProjectContent                  CollectionDiff[models.SalesProjectContent]            `json:"SalesProjectContent"`
	SalesProjectContentAdvancedCondition CollectionDiff[models.SalesProjectAdvancedConditions] `json:"SalesProjectContentAdvancedCondition"`
	SalesProjectItems                    CollectionDiff[models.SalesProjectItems]              `json:"SalesProjectItems"`
	SalesProjectWirings                  CollectionDiff[models.SalesProjectWiring]             `json:"SalesProjectWirings"`
}

type CollectionDiff[T any] struct {
	Added   []T              `json:"Added"`
	Removed []T              `json:"Removed"`
	Updated []UpdatedItem[T] `json:"Updated"`
}

type UpdatedItem[T any] struct {
	Item    T                           `json:"Item"`
	Changes map[string]FieldChange[any] `json:"Changes"`
}

type FieldChange[T any] struct {
	OldValue T `json:"OldValue"`
	NewValue T `json:"NewValue"`
}

type QuotationFieldChanges struct {
	ProjectName          *FieldChange[any] `json:"project_name,omitempty"`
	CustomerID           *FieldChange[any] `json:"customer_id,omitempty"`
	ApplicationID        *FieldChange[any] `json:"application_id,omitempty"`
	PaymentTermsID       *FieldChange[any] `json:"payment_terms_id,omitempty"`
	ShipToID             *FieldChange[any] `json:"ship_to_id,omitempty"`
	BillToID             *FieldChange[any] `json:"bill_to_id,omitempty"`
	ShipTypeID           *FieldChange[any] `json:"ship_type_id,omitempty"`
	Purpose              *FieldChange[any] `json:"purpose,omitempty"`
	Date                 *FieldChange[any] `json:"date,omitempty"`
	ValidityDays         *FieldChange[any] `json:"validity_days,omitempty"`
	ValidUntil           *FieldChange[any] `json:"valid_until,omitempty"`
	Warranty             *FieldChange[any] `json:"warranty,omitempty"`
	AddressTo            *FieldChange[any] `json:"address_to,omitempty"`
	Thru                 *FieldChange[any] `json:"thru,omitempty"`
	GrossSales           *FieldChange[any] `json:"gross_sales,omitempty"`
	VatAmount            *FieldChange[any] `json:"vat_amount,omitempty"`
	NetSales             *FieldChange[any] `json:"net_sales,omitempty"`
	PercentDiscount      *FieldChange[any] `json:"percent_discount,omitempty"`
	DiscountedAmount     *FieldChange[any] `json:"discounted_amount,omitempty"`
	AdditionalDiscounted *FieldChange[any] `json:"additional_discounted,omitempty"`
	CashDiscount         *FieldChange[any] `json:"cash_discount,omitempty"`
	NetAmountDue         *FieldChange[any] `json:"net_amount_due,omitempty"`
	TotalAmountDue       *FieldChange[any] `json:"total_amount_due,omitempty"`
	Contact1             *FieldChange[any] `json:"contact_1,omitempty"`
	Contact2             *FieldChange[any] `json:"contact_2,omitempty"`
	DocumentNo           *FieldChange[any] `json:"document_no,omitempty"`
	VersionNo            *FieldChange[any] `json:"version_no,omitempty"`
	SubVersionNo         *FieldChange[any] `json:"sub_version_no,omitempty"`
	CreatedBy            *FieldChange[any] `json:"created_by,omitempty"`
	FinalRefNo           *FieldChange[any] `json:"final_ref_no,omitempty"`
	IsFinalized          *FieldChange[any] `json:"is_finalized,omitempty"`
	IsProject            *FieldChange[any] `json:"is_project,omitempty"`
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
		// Same table Quick Quote's selected images ride in - a separate JSON key here
		// since these rows are keyed against project items' IDs, not quick quote IDs,
		// even though it's the same underlying table.
		SalesProjectItemsSelectedImages []models.SalesQuotationSelectedImage `json:"sales_project_items_selected_images"`
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
	if err := GetQuotationQuickSelectedImages(&response.SalesProjectItemsSelectedImages, conditions); err != nil {
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

	// Guard: same fix as CreateSalesQuotation (Sales_Quotation_Bug_Report_2026-08-03.md
	// #13) - the Project Quote path has its own separate create function and was
	// equally unprotected. FinalizeProjectQuotation's own client-side duplicate
	// check (against transactionProjectDataTable) was the only thing catching this
	// before, with the exact same staleness/race exposure.
	// Assign the server-side document number BEFORE the uniqueness guard, same as the
	// quick-quote create path - both draw from the one shared tbl_trans_sales_quotation
	// sequence. For a brand-new project quote this replaces the client's guess (which could
	// collide with a quick quote it never loaded); a New Version keeps its number.
	if err := assignNewQuotationDocNo(tx, &body.SalesQuotation); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if body.DocumentNo != "" {
		var existingCount int64
		if err := tx.Model(&models.SalesQuotation{}).
			Where("document_no = ? AND version_no = ? AND sub_version_no = ?",
				body.DocumentNo, body.VersionNo, body.SubVersionNo).
			Count(&existingCount).Error; err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed checking for an existing quotation with this document number")
		}
		if existingCount > 0 {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"a quotation with document number %s (version %s, sub-version %s) already exists",
				body.DocumentNo, body.VersionNo, body.SubVersionNo)
		}
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

	for _, x := range body.SalesProjectAllTabs {

		for _, v := range x.SalesProjectHistory {
			if err := CreateSalesProjectHistory(tx, body.ID, v, at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}

		if err := CreateProjectItemSet(tx, body.ID, &x.SalesProjectItemSet, at); err != nil {
			fmt.Print("KEY ITEM SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		if err := CreateProjectContent(tx, x.SalesProjectItemSet.ItemSetID, x.SalesProjectContent, at); err != nil {
			fmt.Print("KEY CONTENT SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		if err := CreateProjectAdvancedConditions(tx, x.SalesProjectItemSet.ItemSetID, x.SalesProjectContentAdvancedCondition, at); err != nil {
			fmt.Print("KEY ADVCOND SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		for _, v := range x.SalesProjectItems {
			if err := CreateProjectItems(tx, x.SalesProjectItemSet.ItemSetID, v.SalesProjectItems, v.QuickSelectedImage, at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}

		for _, v := range x.SalesProjectWirings {
			if err := CreateProjectWiring(tx, x.SalesProjectItemSet.ItemSetID, v, at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	}

	return body, 0, nil
}

func UpdateSalesProject(c *fiber.Ctx, tx *gorm.DB) (UpdateProjectBody, int, error) {
	var body UpdateProjectBody

	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, fmt.Errorf("cannot bind request: %w", err)
	}

	if body.ID == 0 {
		return body, fiber.StatusBadRequest, errors.New("invalid sales project id")
	}

	at, _ := c.Locals("at").(models.At)

	// ---- UPDATE SALES QUOTATION FIELDS ----
	if err := applyQuotationFieldChanges(tx, body.ID, body.Header.QuotationFields); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("update sales quotation fields: %w", err)
	}

	// ---- UPDATE SALES QUOTATION AT ----
	atData := models.SalesQuotationAt{
		RefId: body.ID,
		At:    at,
	}
	if err := services.DbUpdate(tx, &atData, map[string]interface{}{
		"ref_id": body.ID,
	}); err != nil {
		return body, fiber.StatusInternalServerError, fmt.Errorf("update quotation at: %w", err)
	}

	// ---- MULTIPLIERS ----
	if err := applyMultiplierDiff(tx, body.ID, body.Header.Multipliers, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- HEADER-LEVEL HISTORY (project name/other top fields + multipliers) ----
	if err := applyHistoryDiff(tx, body.ID, body.Header.SalesProjectHistory, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	// ---- TABS ----
	for _, tab := range body.Tabs {

		// ---- ITEM SET ----
		// Resolve this tab's real item_set_id before touching any of its children. For an
		// existing tab, tab.BasedId (sent by the client) already IS the real id. For a
		// brand-new tab, tab.BasedId is just a client-side placeholder - applyItemSetDiff
		// inserts the new tbl_trans_sales_project_item_set row and hands back its real,
		// database-assigned id, which we then use for everything below instead.
		//
		// Previously every child record here (content/advanced conditions/items/wirings)
		// was saved with body.ID (the quotation's own id) as its based_id, which is wrong -
		// based_id on those tables means "which tab/item set", not "which quotation" (see
		// the "SHOULD BE THE TAB # / SET #" comments on the model structs). That silently
		// mis-filed every child record under the quotation id instead of the tab, and a
		// brand-new tab's item set was never reachable at all because nothing resolved its id.
		resolvedItemSetId := tab.BasedId
		if newItemSetId, err := applyItemSetDiff(tx, body.ID, tab.SalesProjectItemSet, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		} else if newItemSetId != 0 {
			resolvedItemSetId = newItemSetId
		}

		// ---- HISTORY ----
		if err := applyHistoryDiff(tx, resolvedItemSetId, tab.SalesProjectHistory, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		// ---- CONTENT ----
		if err := applyContentDiff(tx, resolvedItemSetId, tab.SalesProjectContent, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		// ---- ADVANCED CONDITIONS ----
		if err := applyAdvancedConditionDiff(tx, resolvedItemSetId, tab.SalesProjectContentAdvancedCondition, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		// ---- ITEMS ----
		if err := applyItemsDiff(tx, resolvedItemSetId, tab.SalesProjectItems, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}

		// ---- WIRINGS ----
		if err := applyWiringsDiff(tx, resolvedItemSetId, tab.SalesProjectWirings, at); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
	}

	// NOTE: the "quotation_saved" broadcast deliberately does NOT happen here. This function
	// runs inside the caller's still-open transaction (see UpdateSalesProject in
	// sales_project_handler.go, which calls tx.Commit() after this returns) - broadcasting
	// before that commit could notify another client to refresh before this save is actually
	// durable, racing them into re-reading pre-commit (or, worse, rolled-back) data. The
	// broadcast happens in the handler, after tx.Commit() succeeds.

	return body, fiber.StatusOK, nil
}

func applyQuotationFieldChanges(tx *gorm.DB, id uint, fields QuotationFieldChanges) error {
	updates := map[string]interface{}{}

	if fields.ProjectName != nil {
		updates["project_name"] = fields.ProjectName.NewValue
	}
	if fields.CustomerID != nil {
		updates["customer_id"] = fields.CustomerID.NewValue
	}
	if fields.ApplicationID != nil {
		updates["application_id"] = fields.ApplicationID.NewValue
	}
	if fields.PaymentTermsID != nil {
		updates["payment_terms_id"] = fields.PaymentTermsID.NewValue
	}
	if fields.ShipToID != nil {
		updates["ship_to_id"] = fields.ShipToID.NewValue
	}
	if fields.BillToID != nil {
		updates["bill_to_id"] = fields.BillToID.NewValue
	}
	if fields.ShipTypeID != nil {
		updates["ship_type_id"] = fields.ShipTypeID.NewValue
	}
	if fields.Purpose != nil {
		updates["purpose"] = fields.Purpose.NewValue
	}
	if fields.Date != nil {
		updates["date"] = fields.Date.NewValue
	}
	if fields.ValidityDays != nil {
		updates["validity_days"] = fields.ValidityDays.NewValue
	}
	if fields.ValidUntil != nil {
		updates["valid_until"] = fields.ValidUntil.NewValue
	}
	if fields.Warranty != nil {
		updates["warranty"] = fields.Warranty.NewValue
	}
	if fields.AddressTo != nil {
		updates["address_to"] = fields.AddressTo.NewValue
	}
	if fields.Thru != nil {
		updates["thru"] = fields.Thru.NewValue
	}
	if fields.GrossSales != nil {
		updates["gross_sales"] = fields.GrossSales.NewValue
	}
	if fields.VatAmount != nil {
		updates["vat_amount"] = fields.VatAmount.NewValue
	}
	if fields.NetSales != nil {
		updates["net_sales"] = fields.NetSales.NewValue
	}
	if fields.PercentDiscount != nil {
		updates["percent_discount"] = fields.PercentDiscount.NewValue
	}
	if fields.DiscountedAmount != nil {
		updates["discounted_amount"] = fields.DiscountedAmount.NewValue
	}
	if fields.AdditionalDiscounted != nil {
		updates["additional_discounted"] = fields.AdditionalDiscounted.NewValue
	}
	if fields.CashDiscount != nil {
		updates["cash_discount"] = fields.CashDiscount.NewValue
	}
	if fields.NetAmountDue != nil {
		updates["net_amount_due"] = fields.NetAmountDue.NewValue
	}
	if fields.TotalAmountDue != nil {
		updates["total_amount_due"] = fields.TotalAmountDue.NewValue
	}
	if fields.Contact1 != nil {
		updates["contact_1"] = fields.Contact1.NewValue
	}
	if fields.Contact2 != nil {
		updates["contact_2"] = fields.Contact2.NewValue
	}
	if fields.DocumentNo != nil {
		updates["document_no"] = fields.DocumentNo.NewValue
	}
	if fields.VersionNo != nil {
		updates["version_no"] = fields.VersionNo.NewValue
	}
	if fields.SubVersionNo != nil {
		updates["sub_version_no"] = fields.SubVersionNo.NewValue
	}
	if fields.CreatedBy != nil {
		updates["created_by"] = fields.CreatedBy.NewValue
	}
	if fields.FinalRefNo != nil {
		updates["final_ref_no"] = fields.FinalRefNo.NewValue
	}
	if fields.IsFinalized != nil {
		updates["is_finalized"] = fields.IsFinalized.NewValue
	}
	if fields.IsProject != nil {
		updates["is_project"] = fields.IsProject.NewValue
	}

	if len(updates) == 0 {
		return nil
	}

	if err := tx.Model(&models.SalesQuotation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return err
	}

	// Unlike every other apply*Diff function in this file, this one updates via a raw
	// GORM call (Updates takes a partial map, not a full model struct, so it can't go
	// through services.DbUpdate as-is) - which means it never invalidated the
	// SalesQuotation cache the way the others do via DbUpdate/DbInsert/DbDelete. The
	// database write itself was always correct; the app just kept serving the
	// pre-edit cached copy on every GET until the process restarted, e.g. an edited
	// Application field showing the database's new value in a SQL client but not in
	// the app's own view even after refreshing.
	if err := services.InvalidateCache(services.GetKey(&models.SalesQuotation{}, nil)); err != nil {
		return err
	}
	return services.InvalidateCacheByModel(&models.SalesQuotation{})
}

func applyHistoryDiff(tx *gorm.DB, BasedId uint, diff CollectionDiff[models.SalesProjectHistory], at models.At) error {
	for _, item := range diff.Added {
		item.BasedId = BasedId
		// Same reason as CreateProjectContent's ContentID = 0 - always let the DB assign a
		// fresh id for anything landing in Added, never trust a client-sent HistoryID. The
		// auto-generated Change History entries always send 0 here anyway, but this closes
		// off the same class of collision for any future caller that doesn't.
		item.HistoryID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("history add: %w", err)
		}
	}
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectHistory{}, map[string]interface{}{
			"history_id": item.HistoryID,
		}); err != nil {
			return fmt.Errorf("history remove: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"history_id": entry.Item.HistoryID,
		}); err != nil {
			return fmt.Errorf("history update: %w", err)
		}
	}
	return nil
}

// applyItemSetDiff applies the tab's own item-set-level changes and returns the real,
// database-assigned item_set_id when a new item set was inserted (0 if the tab already
// existed and its item set wasn't itself Added). Callers should fall back to the tab's
// own resolved BasedId when this returns 0.
func applyItemSetDiff(tx *gorm.DB, projectID uint, diff CollectionDiff[models.SalesProjectItemSet], at models.At) (uint, error) {
	var newItemSetId uint
	for _, item := range diff.Added {
		item.BasedId = projectID
		// The client can only ever send a placeholder here for a brand-new tab (there's no
		// real id yet) - never trust it as a real primary key to insert with.
		item.ItemSetID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return 0, fmt.Errorf("item set add: %w", err)
		}
		newItemSetId = item.ItemSetID
	}
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectItemSet{}, map[string]interface{}{
			"item_set_id": item.ItemSetID,
		}); err != nil {
			return 0, fmt.Errorf("item set remove: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		entry.Item.BasedId = projectID
		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"item_set_id": entry.Item.ItemSetID,
		}); err != nil {
			return 0, fmt.Errorf("item set update: %w", err)
		}
	}
	return newItemSetId, nil
}

// syncContentChildren persists a content row's child collections (Final Selection and
// Size Up). Needed because the two branches of applyContentDiff handle them differently
// and neither did so correctly: the ADDED branch leaned on GORM auto-saving the
// associations during DbInsert (which stamps the FK, but also honours whatever id the
// client sent), while the UPDATED branch used DbUpdate, which does not touch associations
// at all - so editing an existing quote silently dropped every final and size-up. That is
// why Size Up never persisted: an existing project quote always takes the UPDATED path.
//
// Rows carrying an id are updated in place, new ones are created with a DB-assigned id,
// and rows the client no longer sends are deleted so removing a candidate sticks.
//
// Both collections are now deduped on the way in and pruned on the way out
// (2026-09-03). Previously only size-ups were pruned, on the reasoning that "the UI
// has no remove action" for finals - but that left the write non-idempotent, so any
// payload that carried the same pump twice silently stored it twice, with nothing
// downstream to catch it. That is exactly what happened to content 14: the client's
// own duplicate guard was broken for reloaded rows (see SalesProjectContentFinal's
// ItemID note), the grid ended up holding each pump twice, and this function wrote
// both copies. The client-side hole is fixed too, but this is the backstop - a
// duplicate can no longer reach the table even if some other caller sends one.
func syncContentChildren(tx *gorm.DB, contentID uint, finals []models.SalesProjectContentFinal, sizeUps []models.SalesProjectSizeUp, at models.At) error {
	// Resolve item ids BEFORE deduping, so the dedupe below can key on the pump
	// itself rather than falling back to comparing model text - two rows naming the
	// same pump are then caught even if one of them arrived without an id.
	resolveChildItemIds(tx, finals, sizeUps)

	finals = dedupeFinals(finals)
	sizeUps = dedupeSizeUps(sizeUps)

	keptFinals := map[uint]bool{}
	for _, v := range finals {
		if v.ID > 0 {
			keptFinals[v.ID] = true
			if err := UpdateProjectContentFinal(tx, v, at, map[string]interface{}{"id": v.ID}); err != nil {
				return fmt.Errorf("content final update: %w", err)
			}
			continue
		}
		if err := CreateProjectContentFinal(tx, contentID, v, at); err != nil {
			return fmt.Errorf("content final add: %w", err)
		}
	}

	var existingFinals []models.SalesProjectContentFinal
	if err := tx.Where("sales_project_content_id = ?", contentID).Find(&existingFinals).Error; err != nil {
		return fmt.Errorf("content final read: %w", err)
	}
	for _, row := range existingFinals {
		if keptFinals[row.ID] {
			continue
		}
		if err := DeleteProjectContentFinal(tx, row, at); err != nil {
			return fmt.Errorf("content final delete: %w", err)
		}
	}

	kept := map[uint]bool{}
	for _, v := range sizeUps {
		if v.ID > 0 {
			kept[v.ID] = true
			if err := UpdateProjectSizeUp(tx, v, at, map[string]interface{}{"id": v.ID}); err != nil {
				return fmt.Errorf("size up update: %w", err)
			}
			continue
		}
		if err := CreateProjectSizeUp(tx, contentID, v, at); err != nil {
			return fmt.Errorf("size up add: %w", err)
		}
	}

	var existing []models.SalesProjectSizeUp
	if err := tx.Where("sales_project_content_id = ?", contentID).Find(&existing).Error; err != nil {
		return fmt.Errorf("size up read: %w", err)
	}
	for _, row := range existing {
		if kept[row.ID] {
			continue
		}
		if err := DeleteProjectSizeUp(tx, row, at); err != nil {
			return fmt.Errorf("size up delete: %w", err)
		}
	}

	return nil
}

// resolveChildItemIds fills in a missing item id on a Size Up / Final row by looking
// its model name up in item setup. Added 2026-09-03: rows written before finals
// carried an item_id have none, and any client that sends only a model name would
// otherwise leave the duplicate check comparing free text.
//
// Only an EXACT, UNAMBIGUOUS match is applied. A model matching two items is left
// alone rather than guessed at - the spec allows non-Calpeda model names to repeat
// across items when their specs differ ("Calpeda model names must be globally
// unique. Other brands may repeat when specs differ"), so picking one would silently
// attach the row to the wrong pump. Leaving it at 0 is safe: the dedupe falls back
// to the model name, which is exactly the information available.
//
// Best-effort throughout - a lookup failure never fails the save, it just leaves the
// id unresolved, because the row itself is still valid without one.
func resolveChildItemIds(tx *gorm.DB, finals []models.SalesProjectContentFinal, sizeUps []models.SalesProjectSizeUp) {
	resolve := func(model string) uint {
		name := strings.TrimSpace(model)
		if name == "" {
			return 0
		}

		var ids []uint
		if err := tx.Model(&models.Item{}).
			Where("item_model = ?", name).
			Limit(2).
			Pluck("id", &ids).Error; err != nil {
			return 0
		}

		// Exactly one match, or nothing - never a guess between two.
		if len(ids) == 1 {
			return ids[0]
		}
		return 0
	}

	for i := range finals {
		if finals[i].ItemID == 0 {
			finals[i].ItemID = resolve(finals[i].Final)
		}
	}

	for i := range sizeUps {
		if sizeUps[i].ItemID == 0 {
			sizeUps[i].ItemID = resolve(sizeUps[i].Model)
		}
	}
}

// dedupeFinals / dedupeSizeUps collapse a payload that names the same pump more than
// once for one content row, keeping the first occurrence. Identity is the pump
// (item_id), falling back to the model name for legacy rows saved before finals
// carried an item id - those have ItemID 0, so keying on it alone would collapse all
// of them into one. A row that already has a database id is always kept as-is: it is
// an existing row being updated, not a fresh duplicate.
func dedupeFinals(rows []models.SalesProjectContentFinal) []models.SalesProjectContentFinal {
	if len(rows) < 2 {
		return rows
	}

	seen := map[string]bool{}
	out := make([]models.SalesProjectContentFinal, 0, len(rows))

	for _, v := range rows {
		key := finalIdentity(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}

	return out
}

func finalIdentity(v models.SalesProjectContentFinal) string {
	if v.ID > 0 {
		return fmt.Sprintf("id:%d", v.ID)
	}
	if v.ItemID > 0 {
		return fmt.Sprintf("item:%d", v.ItemID)
	}
	return "model:" + strings.TrimSpace(strings.ToUpper(v.Final))
}

func dedupeSizeUps(rows []models.SalesProjectSizeUp) []models.SalesProjectSizeUp {
	if len(rows) < 2 {
		return rows
	}

	seen := map[string]bool{}
	out := make([]models.SalesProjectSizeUp, 0, len(rows))

	for _, v := range rows {
		var key string
		switch {
		case v.ID > 0:
			key = fmt.Sprintf("id:%d", v.ID)
		case v.ItemID > 0:
			key = fmt.Sprintf("item:%d", v.ItemID)
		default:
			key = "model:" + strings.TrimSpace(strings.ToUpper(v.Model))
		}

		if seen[key] {
			continue
		}
		seen[key] = true
		out = append(out, v)
	}

	return out
}

func applyContentDiff(tx *gorm.DB, BasedId uint, diff CollectionDiff[models.SalesProjectContent], at models.At) error {
	if diff.Added == nil && diff.Removed == nil && diff.Updated == nil {
		return nil
	}

	// ---- ADDED ----
	for _, item := range diff.Added {
		item.BasedId = BasedId
		// Same reason as CreateProjectContent's ContentID = 0 - a "new" content row can still
		// carry a stale/leftover ContentID from the client (e.g. a UI control that wasn't
		// cleared for a brand-new tab), which would force GORM into an explicit
		// IDENTITY_INSERT with that id and risk a PRIMARY KEY collision. Always let the DB
		// assign a fresh id for anything landing in Added.
		item.ContentID = 0

		// Detach before insert: GORM would auto-save these associations using the ids the
		// client sent, which risks an IDENTITY_INSERT collision. syncContentChildren below
		// creates them with DB-assigned ids and the parent's real ContentID.
		addFinals := item.SalesProjectContentFinal
		addSizeUps := item.SalesProjectSizeUp
		item.SalesProjectContentFinal = nil
		item.SalesProjectSizeUp = nil

		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("content add: %w", err)
		}

		if err := syncContentChildren(tx, item.ContentID, addFinals, addSizeUps, at); err != nil {
			return err
		}

		atRecord := models.SalesProjectContentAt{
			RefID:                      item.ContentID,
			SalesProjectContentContent: item.SalesProjectContentContent,
			At:                         at,
		}
		if err := services.DbInsert(tx, &atRecord); err != nil {
			return fmt.Errorf("content add at: %w", err)
		}
	}

	// ---- REMOVED ----
	for _, item := range diff.Removed {
		// Children first. Size Up and Final Selection both carry a foreign key to this
		// row (fk_tbl_trans_sales_project_content_sales_project_size_up and its Final
		// counterpart), so deleting the parent on its own is refused outright by SQL
		// Server - "The DELETE statement conflicted with the REFERENCE constraint" -
		// and the whole save fails. Each child goes through its own delete helper so
		// the removal still lands in the audit tables rather than vanishing.
		var sizeUps []models.SalesProjectSizeUp
		if err := tx.Where("sales_project_content_id = ?", item.ContentID).Find(&sizeUps).Error; err != nil {
			return fmt.Errorf("content remove (size up read): %w", err)
		}
		for _, row := range sizeUps {
			if err := DeleteProjectSizeUp(tx, row, at); err != nil {
				return fmt.Errorf("content remove (size up): %w", err)
			}
		}

		var finals []models.SalesProjectContentFinal
		if err := tx.Where("sales_project_content_id = ?", item.ContentID).Find(&finals).Error; err != nil {
			return fmt.Errorf("content remove (final read): %w", err)
		}
		for _, row := range finals {
			if err := DeleteProjectContentFinal(tx, row, at); err != nil {
				return fmt.Errorf("content remove (final): %w", err)
			}
		}

		if err := services.DbDelete(tx, &models.SalesProjectContent{}, map[string]interface{}{
			"content_id": item.ContentID,
		}); err != nil {
			return fmt.Errorf("content remove: %w", err)
		}
	}

	// ---- UPDATED ----
	for _, entry := range diff.Updated {

		println(entry.Item.ItemDesignation, " ID=", entry.Item.ContentID)

		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"content_id": entry.Item.ContentID,
		}); err != nil {
			return fmt.Errorf("content update: %w", err)
		}

		// DbUpdate does not touch associations, so without this the finals and size-ups
		// on an edited tab were discarded on every save.
		if err := syncContentChildren(tx, entry.Item.ContentID, entry.Item.SalesProjectContentFinal, entry.Item.SalesProjectSizeUp, at); err != nil {
			return err
		}

		atRecord := models.SalesProjectContentAt{
			RefID:                      entry.Item.ContentID,
			SalesProjectContentContent: entry.Item.SalesProjectContentContent,
			At:                         at,
		}
		if err := services.DbUpdate(tx, &atRecord, map[string]interface{}{
			"ref_id": entry.Item.ContentID,
		}); err != nil {
			return fmt.Errorf("content update at: %w", err)
		}
	}

	return nil
}

func applyAdvancedConditionDiff(tx *gorm.DB, basedId uint, diff CollectionDiff[models.SalesProjectAdvancedConditions], at models.At) error {
	for _, item := range diff.Added {
		item.BasedId = basedId
		// Same reason as CreateProjectContent's ContentID = 0 - always let the DB assign a
		// fresh id for anything landing in Added, never trust a client-sent ConditionsID.
		item.ConditionsID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("advanced condition add: %w", err)
		}
	}
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectAdvancedConditions{}, map[string]interface{}{
			"conditions_id": item.ConditionsID,
		}); err != nil {
			return fmt.Errorf("advanced condition remove: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"conditions_id": entry.Item.ConditionsID,
		}); err != nil {
			return fmt.Errorf("advanced condition update: %w", err)
		}
	}
	return nil
}

func applyItemsDiff(tx *gorm.DB, basedId uint, diff CollectionDiff[models.SalesProjectItems], at models.At) error {
	for _, item := range diff.Added {
		item.BasedId = basedId
		// Same reason as CreateProjectContent's ContentID = 0 - always let the DB assign a
		// fresh id for anything landing in Added, never trust a client-sent ItemsID.
		item.ItemsID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("item add: %w", err)
		}
	}
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectItems{}, map[string]interface{}{
			"items_id": item.ItemsID,
		}); err != nil {
			return fmt.Errorf("item remove: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"items_id": entry.Item.ItemsID,
		}); err != nil {
			return fmt.Errorf("item update: %w", err)
		}
	}
	return nil
}

func applyWiringsDiff(tx *gorm.DB, basedId uint, diff CollectionDiff[models.SalesProjectWiring], at models.At) error {
	for _, item := range diff.Added {
		item.BasedId = basedId
		// Same reason as CreateProjectContent's ContentID = 0 - always let the DB assign a
		// fresh id for anything landing in Added, never trust a client-sent ID.
		item.ID = 0
		if err := services.DbInsert(tx, &item); err != nil {
			return fmt.Errorf("wiring add: %w", err)
		}
	}
	// "id", not "wiring_id" (fixed 2026-09-04). tbl_trans_sales_project_wiring's
	// primary key is plain "id" - SalesProjectWiring declares `ID uint gorm:"primarykey"`
	// with no column override - so every edit to an existing wiring row failed outright
	// with "Invalid column name 'wiring_id'", and the whole save rolled back with it.
	// Delete had the identical typo and would have failed the same way. Every other
	// applier in this file names its table's real key (multiplier_id, history_id,
	// item_set_id, content_id, conditions_id, items_id) - wiring was the only one that
	// invented a name.
	for _, item := range diff.Removed {
		if err := services.DbDelete(tx, &models.SalesProjectWiring{}, map[string]interface{}{
			"id": item.ID,
		}); err != nil {
			return fmt.Errorf("wiring remove: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		if err := services.DbUpdate(tx, &entry.Item, map[string]interface{}{
			"id": entry.Item.ID,
		}); err != nil {
			return fmt.Errorf("wiring update: %w", err)
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

	for _, x := range body.SalesProjectAllTabs {

		if err := CreateProjectItemSet(tx, x.SalesProjectItemSet.BasedId, &x.SalesProjectItemSet, at); err != nil {
			fmt.Print("KEY ITEM SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		if err := CreateProjectContent(tx, x.SalesProjectItemSet.ItemSetID, x.SalesProjectContent, at); err != nil {
			fmt.Print("KEY CONTENT SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		if err := CreateProjectAdvancedConditions(tx, x.SalesProjectItemSet.ItemSetID, x.SalesProjectContentAdvancedCondition, at); err != nil {
			fmt.Print("KEY ADVCOND SET::", err)
			return body, fiber.StatusInternalServerError, err
		}

		for _, v := range x.SalesProjectItems {
			if err := CreateProjectItems(tx, x.SalesProjectItemSet.ItemSetID, v.SalesProjectItems, v.QuickSelectedImage, at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}

		for _, v := range x.SalesProjectWirings {
			if err := CreateProjectWiring(tx, x.SalesProjectItemSet.ItemSetID, v, at); err != nil {
				return body, fiber.StatusInternalServerError, err
			}
		}
	}

	return body, 0, nil
}

func UpdateProjectContents(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (models.SalesProjectContent, int, error) {
	var body models.SalesProjectContent
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}

	if err := UpdateProjectContent(tx, body, at, conditions); err != nil {
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

func UpdateProjectItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (NewProjectItem, int, error) {
	var body NewProjectItem
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
