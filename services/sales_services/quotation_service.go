package sales_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

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

// recomputeQuickQuoteTotals is the server's own authority on a Quick Quote's
// totals - §8.2's formulas, computed independently of whatever the client
// sent, then overwritten onto quotation before it's persisted. This is
// deliberately "recompute and overwrite," not "recompute and reject on
// mismatch": the client's own version already exists purely for live UI
// feedback as the user edits (ComputeReferenceNonHierarchy/ComputeFooterTotals
// in Quotation.cs) and is never itself persisted as authoritative - comparing
// with a tolerance and rejecting on drift would just risk false rejections
// from ordinary floating-point/rounding differences between the two
// implementations, for no real benefit over simply always trusting one
// server-side computation.
//
// additional_discounted and cash_discount are NOT recomputed here - §8.2
// itself defines both as genuine inputs ("additional discount = % input",
// "cash discount = peso input"), not values derivable from line items. They
// are trusted as given and used as-is in the formulas below.
//
// §8.2 also excludes BOM child rows from every total ("gross sales is
// computed, not summed from a column" - only top-level items count,
// CLAUDE.md invariant #8). Quick Quote's own hierarchy marker is
// ReferenceCode containing "." (a child, e.g. "1.1") - Quotation.cs's own
// ComputeReferenceNonHierarchy skips these on the client the same way.
//
// VAT rate: this system uses one flat, company-wide rate (Company Setup,
// backlog #18 - the user's own explicit choice over a per-quotation Tax Setup
// lookup), fetched fresh here rather than trusting anything the client sent.
// Known, accepted limitation: CLAUDE.md invariant #6 calls for a tax rate to
// be frozen at a document's creation and never re-derived - there is no field
// on this model to persist which rate a given save actually used, so a plain
// edit made after the company rate changes will recompute VAT at the NEW
// rate, not whatever rate was in effect when the quotation was first created.
// This is a pre-existing gap (the current, unprotected code doesn't freeze
// anything either - it just trusts whatever the client sends), and finalize
// itself (backlog #11's own is_finalized lock) is what makes a quotation's
// numbers actually permanent - once finalized, no further edit is possible
// to be inconsistent with. Worth revisiting if a real "frozen rate" field is
// ever added.
func recomputeQuickQuoteTotals(tx *gorm.DB, quotation *models.SalesQuotation, lines []SalesQuotationQuickWithImages) error {
	var grossSales, netSales float64

	for _, line := range lines {
		// Child/component row (e.g. "1.1") - excluded from every total, same
		// as ComputeReferenceNonHierarchy's own client-side skip.
		if strings.Contains(line.ReferenceCode, ".") {
			continue
		}

		undiscountedPrice := line.UnitPrice
		if line.ListPrice > 0 {
			undiscountedPrice = line.ListPrice
		}

		grossSales += float64(line.Qty) * undiscountedPrice
		netSales += line.LineTotal
	}

	var percentDiscount float64
	if grossSales != 0 {
		percentDiscount = ((grossSales - netSales) / grossSales) * 100
	}
	discountedAmount := grossSales - netSales

	var company models.CompanyModel
	if err := tx.First(&company, 1).Error; err != nil {
		return errors.New("failed fetching company VAT rate for totals recompute")
	}
	vatRate := company.VatRatePercent / 100

	netOfVat := netSales - quotation.AdditionalDiscounted
	vatAmount := netOfVat * vatRate
	netAmount := netOfVat + vatAmount
	totalAmountDue := netAmount - quotation.CashDiscount

	quotation.GrossSales = grossSales
	quotation.NetSales = netSales
	quotation.PercentDiscount = percentDiscount
	quotation.DiscountedAmount = discountedAmount
	quotation.VatAmount = vatAmount
	quotation.NetAmountDue = netAmount
	quotation.TotalAmountDue = totalAmountDue

	return nil
}

func CreateSalesQuotation(c *fiber.Ctx, tx *gorm.DB) (CreateBody, int, error) {
	var body CreateBody
	if err := c.BodyParser(&body); err != nil {
		fmt.Println("pt 1", err)
		fmt.Println("ERR", body)
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Guard: Sales_Quotation_Bug_Report_2026-08-03.md #13 - nothing enforced
	// uniqueness of (document_no, version_no, sub_version_no), so a stale
	// client-side list (Quotation.cs's own allTransactionList/
	// transactionProjectDataTable duplicate check, which is the ONLY thing
	// that ever caught this before) or a retried/raced request could create
	// two rows for what should be one specific version - most visibly,
	// clicking Finalize a second time on the same draft (a stale local list
	// after the first finalize hadn't reloaded yet, a second window, etc.)
	// silently created a second "FQ#..." copy with no error at all.
	// Confirmed reproducible directly against this exact function's own
	// insert shape before adding this. version_no/sub_version_no are matched
	// as-is (including both empty) so this doesn't block New Version, which
	// legitimately reuses the same document_no with a different version_no.
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

	// Sales_Quotation_Bug_Report_2026-08-03.md #9 - gross/net/VAT/discount/line
	// totals were trusted verbatim from the client with no recomputation, so a
	// buggy client could persist a quotation whose totals don't match its own
	// line items. Scoped to Quick Quote for now - Project Quote's totals come
	// from a hierarchical BOM parent/child structure (ComputeByReferenceHierarchy
	// in Quotation.cs) that's a different, larger computation to replicate
	// server-side; not done in this pass.
	if !body.IsProject {
		if err := recomputeQuickQuoteTotals(tx, &body.SalesQuotation, body.SalesQuotationQuickWithImages); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
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

	// Guard: a Finalized Quote is immutable (CLAUDE.md invariant #2 / §5.2 - "an
	// exact, uneditable copy of the SQ at finalization"). The only real
	// protection against editing one used to be that the Sales app hides its
	// own Edit button once isFinalized is true (bind(), Quotation.cs) - nothing
	// stopped a direct call to this endpoint. Checked against the row's
	// CURRENT (pre-update) state, not the incoming body's own IsFinalized -
	// that's what lets the finalize action itself (false -> true, this same
	// endpoint) through while blocking any edit attempt after it.
	var existingQuotation models.SalesQuotation
	if err := tx.First(&existingQuotation, body.SalesQuotation.ID).Error; err != nil {
		return body, fiber.StatusNotFound, errors.New("quotation not found")
	}
	if existingQuotation.IsFinalized {
		return body, fiber.StatusForbidden, errors.New("cannot edit a finalized quotation")
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

	// Sales_Quotation_Bug_Report_2026-08-03.md #9 - same recompute as
	// CreateSalesQuotation, applied here too since this endpoint also handles
	// a plain update as well as finalize (see recomputeQuickQuoteTotals's own
	// comment for why this is "recompute and overwrite," not
	// "recompute and reject on mismatch," and why Project Quote is out of
	// scope for this pass).
	if !body.IsProject {
		if err := recomputeQuickQuoteTotals(tx, &body.SalesQuotation, body.SalesQuotationQuickWithImages); err != nil {
			return body, fiber.StatusInternalServerError, err
		}
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

// RequestQuotationForEngr is §3.2/§6.3's "REQUEST FOR ENGR." action - previously a
// pure client-side stub (btn_request_for_engr_Click just opened an unrelated test
// form) with no backing field anywhere. Makes the per-quote-to-a-specific-engineer
// grant §3.2 describes explicit, rather than the engineering red box/Sales
// Quotation List inferring "sent to engineering" implicitly from "has a project
// name and at least one wiring row" (see vw_get_engineering_redbox_quotation_list.sql),
// which fires with no deliberate action and isn't scoped to any one engineer.
// Phase 4 item 4.1.
func RequestQuotationForEngr(tx *gorm.DB, quotationId uint, engrId uint, at models.At) (*models.SalesQuotation, int, error) {
	if engrId == 0 {
		return nil, fiber.StatusBadRequest, errors.New("engr_id is required - which engineer is this being requested to?")
	}

	var quotation models.SalesQuotation
	if err := tx.First(&quotation, quotationId).Error; err != nil {
		return nil, fiber.StatusNotFound, errors.New("quotation not found")
	}

	var engr models.User
	if err := tx.First(&engr, engrId).Error; err != nil {
		return nil, fiber.StatusBadRequest, errors.New("engineer not found")
	}

	quotation.IsRequestedForEngr = true
	quotation.RequestedEngrId = engrId
	quotation.RequestedEngrName = strings.TrimSpace(engr.FirstName + " " + engr.LastName)
	quotation.RequestedForEngrDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &quotation, map[string]interface{}{"id": quotation.ID}); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed requesting quotation for engineering")
	}

	atdata := models.SalesQuotationAt{
		RefId:                  quotation.ID,
		SalesQuotationContent:  quotation.SalesQuotationContent,
		At:                     at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	return &quotation, fiber.StatusOK, nil
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
