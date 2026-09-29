package sales_services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
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

// assignNewQuotationDocNo gives a brand-new quotation its document number on the server,
// where it is the only place that can see the WHOLE tbl_trans_sales_quotation sequence.
//
// Quick and project quotes share one document_no sequence in that table, but the two
// client screens each load only their own half (GetSalesQuotations filters out projects;
// GetProjects returns only projects). So the client's own DocumentIncrementer computed the
// next number from a partial list and picked one already taken by the other kind - e.g. a
// new Quick Quote landing on Q#0003 when Q#0003 already existed as a project quote.
//
// This assigns only for a genuinely new document - version 1 / sub-version 0. A NEW VERSION
// (version_no >= 2) deliberately reuses its existing document_no, so it is left untouched.
// Finalization does not pass through here at all; it is an UpdateFinalizeQuote path.
//
// SELECT ... WITH (UPDLOCK, HOLDLOCK) takes a range lock for the life of the caller's
// transaction, so two quotes being created at the same instant serialize here instead of
// both reading the same MAX and racing onto the same number (which the
// UQ_sales_quotation_doc_version unique index would then reject on the loser). document_no
// is stored with its "Q#"/"FQ#" prefix, so the numeric part is taken from after the '#'.
func assignNewQuotationDocNo(tx *gorm.DB, sq *models.SalesQuotation) error {
	// Finalization is NOT a fresh number. It keeps its source quote's number and only
	// swaps the prefix (Q#0005 -> FQ#0005), which the client already baked into DocumentNo.
	// A finalize also arrives as version 1 (so isFirstVersion below can't tell it apart from
	// a brand-new draft), which is exactly how this used to clobber FQ#0005 into a fresh
	// Q#000n. Detect it by is_finalized or the FQ# prefix and leave the number untouched.
	if sq.IsFinalized || strings.HasPrefix(strings.ToUpper(strings.TrimSpace(sq.DocumentNo)), "FQ#") {
		return nil
	}

	if !isFirstVersion(sq.VersionNo, sq.SubVersionNo) {
		return nil // a new version keeps the document_no the client sent
	}

	var maxNum int
	err := tx.Raw(`
		SELECT ISNULL(MAX(TRY_CONVERT(int,
			CASE WHEN CHARINDEX('#', document_no) > 0
			     THEN SUBSTRING(document_no, CHARINDEX('#', document_no) + 1, 50)
			     ELSE document_no END)), 0)
		FROM tbl_trans_sales_quotation WITH (UPDLOCK, HOLDLOCK)
	`).Scan(&maxNum).Error
	if err != nil {
		return fmt.Errorf("failed computing next quotation document number: %w", err)
	}

	// A brand-new quote is always a draft "Q#"; only finalization (elsewhere) mints "FQ#".
	sq.DocumentNo = fmt.Sprintf("Q#%04d", maxNum+1)
	return nil
}

// isFirstVersion reports whether these version fields describe a brand-new document rather
// than a later version of an existing one. Treats empty as the first (the client sends "1"
// for a new quote and increments for New Version, but empty is defended against too).
func isFirstVersion(versionNo, subVersionNo string) bool {
	v := strings.TrimSpace(versionNo)
	sv := strings.TrimSpace(subVersionNo)
	return (v == "" || v == "1") && (sv == "" || sv == "0")
}

// quotationListShape is the wire shape GET /sales/quotation has always
// returned: headers + quick lines + selected images. GET
// /sales/quotation/latest and GET /sales/quotation/versions return the same
// shape (narrowed to latest versions / one document) so the WinForms client
// binds any of them with its existing SalesQuotationList model.
type quotationListShape struct {
	SalesQuotation               []models.SalesQuotation
	SalesQuotationQuick          []models.SalesQuotationQuick
	SalesQuotationSelectedImages []models.SalesQuotationSelectedImage
}

func newQuotationListShape() quotationListShape {
	return quotationListShape{
		SalesQuotation:               []models.SalesQuotation{},
		SalesQuotationQuick:          []models.SalesQuotationQuick{},
		SalesQuotationSelectedImages: []models.SalesQuotationSelectedImage{},
	}
}

// scopeQuotationChildren fills quick lines + selected images for exactly the
// given header ids. The full tables are read through the same cached DbGet
// keys the all-versions endpoint uses (so no new cache-invalidation surface),
// and the narrowing happens in API memory - the VPN transfer to the client is
// what shrinks, which is the actual bottleneck.
func scopeQuotationChildren(shape *quotationListShape, headerIDs map[uint]bool) error {
	var quicks []models.SalesQuotationQuick
	if err := GetSalesQuotationQuicks(&quicks, nil); err != nil {
		return err
	}

	var images []models.SalesQuotationSelectedImage
	if err := GetQuotationQuickSelectedImages(&images, nil); err != nil {
		return err
	}

	lineIDs := make(map[uint]bool, len(quicks))
	for _, q := range quicks {
		if headerIDs[q.BasedId] {
			shape.SalesQuotationQuick = append(shape.SalesQuotationQuick, q)
			lineIDs[q.ID] = true
		}
	}
	for _, img := range images {
		if lineIDs[img.QuotationQuickId] {
			shape.SalesQuotationSelectedImages = append(shape.SalesQuotationSelectedImages, img)
		}
	}

	return nil
}

// invalidateLatestQuotationCache clears the vw_latest_sales_quotation read
// cache. DbInsert/DbUpdate invalidate by written model only
// (model:SalesQuotation*), so without this a save stays invisible on GET
// /sales/quotation/latest until the hour cache expires - the same gap
// RequestQuotationForEngr already closes for the engineering list view.
func invalidateLatestQuotationCache() error {
	InvalidateQuotationPipelineViews()
	return services.InvalidateCache(services.GetKey(models.LatestQuotationView{}, nil))
}

// InvalidateQuotationPipelineViews clears the two other lists built on the quotation rows:
// Opportunities (vw_get_sales_opportunities) and BPI Finance's PENDING table
// (vw_get_bpi_finance_pending, which mirrors it - spec 4.7). Each is cached as its own type,
// so the model:SalesQuotation* clear a quotation write does never reached them: a quote
// saved, versioned or finalized left both showing the old list for up to the hour DbGet
// caches for, and only an edit on the Opportunities screen itself refreshed it
// (user-reported 2026-09-25: "the opportunities module is not reflecting").
//
// Called from invalidateLatestQuotationCache, which every quotation write already calls, and
// again after commit from InvalidateEngineeringQuotationListCache - a read that lands between
// the first clear and the commit would otherwise put the pre-commit list straight back.
// Best-effort: the write has succeeded either way.
func InvalidateQuotationPipelineViews() {
	for _, view := range []interface{}{&models.OpportunityView{}, &models.BpiFinancePendingView{}} {
		if err := services.InvalidateCacheByModel(view); err != nil {
			fmt.Println("quotation list cache invalidation:", err)
		}
	}
}

// InvalidateEngineeringQuotationListCache clears the Engineering Sales Quotation
// List's own cached read.
//
// That list is served by DbGet on EngineeringRedboxQuotationListView, bound to
// vw_get_engineering_redbox_quotation_list - a DIFFERENT cached type from
// SalesQuotation, sitting on the same tbl_trans_sales_quotation row. DbUpdate's
// InvalidateCacheByModel only ever clears model:SalesQuotation* keys, so a write
// that changes what the view returns leaves the engineer's list serving its cached
// copy for the full hour DbGet caches for (db_service.go).
//
// Exported and called from the handlers rather than from the services below,
// because it must run AFTER tx.Commit(). Clearing the key while the transaction is
// still open lets a concurrent read repopulate it from pre-commit data, which puts
// the stale list straight back - the same reasoning as the "quotation_saved"
// broadcast, which is also deliberately post-commit.
//
// Best-effort at the call sites: the write itself has already succeeded and been
// committed by then, so a Redis failure must not turn a saved quotation into a
// failed response.
func InvalidateEngineeringQuotationListCache() error {
	InvalidateQuotationPipelineViews()
	return services.InvalidateCache(services.GetKey(models.EngineeringRedboxQuotationListView{}, nil))
}

// stripQuotationPrefix mirrors the WinForms DocumentNo.Strip helper for
// quotation prefixes: the client compares prefix-insensitively
// (NormalizeDocumentNo), so "0005", "Q#0005" and "FQ#0005" must all resolve
// to the same document.
func stripQuotationPrefix(docNo string) string {
	v := strings.TrimSpace(docNo)
	for {
		u := strings.ToUpper(v)
		if strings.HasPrefix(u, "FQ#") {
			v = strings.TrimSpace(v[3:])
			continue
		}
		if strings.HasPrefix(u, "Q#") {
			v = strings.TrimSpace(v[2:])
			continue
		}
		return v
	}
}

// Phase 2 - paged headers + lazy detail. The module list used to download
// every header with every line and image (then every version of them); now it
// pages headers only and pulls one record's children on selection.
//
// quickOnlyWhere keeps the quick/project split server-side. It mirrors the old
// Go filter (ProjectName == "") exactly, NULL included - a plain
// `project_name = ''` would silently drop NULL rows the client used to see.
const quickOnlyWhere = "(project_name IS NULL OR project_name = '')"

const quotationHeaderPageSize = 50

// GetQuotationHeadersPaged serves the module list: one page of latest quick
// headers, newest first, no children. Keyset on id (never OFFSET - ids are
// indexed and stable under inserts), mirroring GetItemsPaged:
//
//	after  - the last id of the page just shown (NEXT >>, older)
//	before - the first id of it (<< PREV, newer)
//	at     - open the page starting at this id (a search hit, one just saved)
//
// All three unset means the first page. Deliberately uncached, like
// GetItemsPaged: an indexed keyset window plus counts is cheap, and page keys
// would need invalidating on every save.
func GetQuotationHeadersPaged(after, before, at int) (quotationListShape, utils.PaginationMeta, int, error) {
	shape := newQuotationListShape()
	pagination := utils.PaginationMeta{PageSize: quotationHeaderPageSize}

	var latest []models.LatestQuotationView
	query := initializers.DB.Model(&models.LatestQuotationView{}).Where(quickOnlyWhere)
	reversed := false

	switch {
	case before > 0:
		query = query.Where("id > ?", before).Order("id ASC")
		reversed = true
	case after > 0:
		query = query.Where("id < ?", after).Order("id DESC")
	case at > 0:
		query = query.Where("id <= ?", at).Order("id DESC")
	default:
		query = query.Order("id DESC")
	}

	if err := query.Limit(quotationHeaderPageSize).Find(&latest).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed getting quotation headers")
	}

	// A PREV page is read back to front; put it the right way round before
	// anything else sees it.
	if reversed {
		for i, j := 0, len(latest)-1; i < j; i, j = i+1, j-1 {
			latest[i], latest[j] = latest[j], latest[i]
		}
	}

	var total int64
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(quickOnlyWhere).Count(&total).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed counting quotation headers")
	}
	pagination.Total = total
	pagination.TotalPages = int((total + int64(quotationHeaderPageSize) - 1) / int64(quotationHeaderPageSize))

	if len(latest) == 0 {
		return shape, pagination, 0, nil
	}

	// Indexed counts over the view's ids - cheap, and exact in both directions.
	var newer, older int64
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(quickOnlyWhere).Where("id > ?", latest[0].ID).Count(&newer).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed checking for a newer header page")
	}
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(quickOnlyWhere).Where("id < ?", latest[len(latest)-1].ID).Count(&older).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed checking for an older header page")
	}
	pagination.HasPrev = newer > 0
	pagination.HasNext = older > 0
	pagination.Page = int(newer)/quotationHeaderPageSize + 1

	for _, v := range latest {
		shape.SalesQuotation = append(shape.SalesQuotation, v.SalesQuotation)
	}

	return shape, pagination, 0, nil
}

// GetQuotationDetail returns ONE header with only its own lines and images,
// for lazy loading on row selection. 404 when the id is unknown.
func GetQuotationDetail(id int) (quotationListShape, int, error) {
	shape := newQuotationListShape()

	if id <= 0 {
		return shape, fiber.StatusBadRequest, errors.New("missing or invalid quotation id")
	}

	var header models.SalesQuotation
	if err := services.DbGetNoCache(&header, map[string]interface{}{"id": id}); err != nil {
		return shape, fiber.StatusInternalServerError, errors.New("failed getting quotation")
	}
	if header.ID == 0 {
		return shape, fiber.StatusNotFound, errors.New("quotation not found")
	}
	shape.SalesQuotation = append(shape.SalesQuotation, header)

	ids := map[uint]bool{header.ID: true}
	if err := scopeQuotationChildrenNoCache(&shape, ids); err != nil {
		return shape, fiber.StatusInternalServerError, err
	}

	return shape, 0, nil
}

// scopeQuotationChildrenNoCache is scopeQuotationChildren without the shared
// full-table cache keys: a single record's children are two indexed seeks
// (based_id, then quotation_quick_id IN), cheap enough to read straight from
// the database every time, and per-id cache keys would churn on each save.
func scopeQuotationChildrenNoCache(shape *quotationListShape, headerIDs map[uint]bool) error {
	headerList := make([]uint, 0, len(headerIDs))
	for id := range headerIDs {
		headerList = append(headerList, id)
	}

	var quicks []models.SalesQuotationQuick
	if err := services.DbGetNoCache(&quicks, map[string]interface{}{"based_id": headerList}); err != nil {
		return err
	}

	lineIDs := make([]uint, 0, len(quicks))
	seen := make(map[uint]bool, len(quicks))
	for _, q := range quicks {
		shape.SalesQuotationQuick = append(shape.SalesQuotationQuick, q)
		if !seen[q.ID] {
			seen[q.ID] = true
			lineIDs = append(lineIDs, q.ID)
		}
	}

	if len(lineIDs) == 0 {
		return nil
	}

	var images []models.SalesQuotationSelectedImage
	if err := services.DbGetNoCache(&images, map[string]interface{}{"quotation_quick_id": lineIDs}); err != nil {
		return err
	}
	shape.SalesQuotationSelectedImages = append(shape.SalesQuotationSelectedImages, images...)

	return nil
}

// SearchQuotations backs the module Search: flat matching quick headers, 20
// to a page, no children - the picked record reopens through the headers'
// `at`. A blank term matches nothing (never the whole table).
func SearchQuotations(term string, page int) (quotationListShape, utils.PaginationMeta, int, error) {
	const searchPageSize = 20

	shape := newQuotationListShape()
	pagination := utils.PaginationMeta{PageSize: searchPageSize, Page: page}

	term = strings.TrimSpace(term)
	if term == "" {
		return shape, pagination, 0, nil
	}
	if page < 1 {
		page = 1
		pagination.Page = 1
	}

	base := initializers.DB.Model(&models.SalesQuotation{}).
		Where(quickOnlyWhere).
		Where("document_no LIKE ?", "%"+term+"%")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed searching quotations")
	}
	pagination.Total = total
	pagination.TotalPages = int((total + int64(searchPageSize) - 1) / int64(searchPageSize))
	pagination.HasPrev = page > 1
	pagination.HasNext = page < pagination.TotalPages

	if err := base.Order("id DESC").
		Offset((page - 1) * searchPageSize).
		Limit(searchPageSize).
		Find(&shape.SalesQuotation).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed searching quotations")
	}

	return shape, pagination, 0, nil
}

func GetLatestQuotations(conditions map[string]interface{}) (interface{}, int, error) {
	shape := newQuotationListShape()

	// One row per document out of SQL (ROW_NUMBER partition in the view),
	// instead of every historical version - the client used to download all
	// of them and dedupe in memory (Quotation.cs fetchQuotationDetails).
	var latest []models.LatestQuotationView
	if err := services.DbGet(&latest, conditions); err != nil {
		return shape, fiber.StatusInternalServerError, errors.New("failed getting latest quotations")
	}

	ids := make(map[uint]bool, len(latest))
	for _, v := range latest {
		if v.ProjectName == "" {
			shape.SalesQuotation = append(shape.SalesQuotation, v.SalesQuotation)
			ids[v.ID] = true
		}
	}

	if err := scopeQuotationChildren(&shape, ids); err != nil {
		return shape, fiber.StatusInternalServerError, err
	}

	return shape, 0, nil
}

// GetQuotationVersions returns every version of ONE quick quotation document
// (all version_no/sub_version_no rows for the bare document number) with only
// that document's lines and images. It backs the WinForms version-history
// viewer (VersionModal) and the open-by-document flows, which previously
// downloaded the full quotation list to look at a single document.
func GetQuotationVersions(documentNo string) (interface{}, int, error) {
	shape := newQuotationListShape()

	bare := stripQuotationPrefix(documentNo)
	if bare == "" {
		return shape, fiber.StatusBadRequest, errors.New("missing document number")
	}

	// Same shared cache key as the all-versions endpoint - no new
	// invalidation surface; the narrowing is in API memory only.
	var all []models.SalesQuotation
	if err := services.DbGet(&all, nil); err != nil {
		return shape, fiber.StatusInternalServerError, errors.New("failed getting quotation versions")
	}

	ids := make(map[uint]bool)
	for _, q := range all {
		if q.ProjectName == "" && stripQuotationPrefix(q.DocumentNo) == bare {
			shape.SalesQuotation = append(shape.SalesQuotation, q)
			ids[q.ID] = true
		}
	}

	if len(shape.SalesQuotation) == 0 {
		return shape, fiber.StatusNotFound, errors.New("quotation not found")
	}

	if err := scopeQuotationChildren(&shape, ids); err != nil {
		return shape, fiber.StatusInternalServerError, err
	}

	return shape, 0, nil
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
	// Assign the server-side document number BEFORE the uniqueness guard below: for a
	// brand-new quote this replaces the client's (possibly colliding) guess with a fresh
	// number, so the guard then validates the number that will actually be inserted. A New
	// Version is left as-is and the guard checks it unchanged.
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

	// Changed 2026-09-03 (user decision): REQUEST FOR ENGR. can now be checked
	// before a quotation is ever saved (Quotation.cs's btn_request_for_engr_Click
	// just flips the checkbox locally in that case), and is_requested_for_engr rides
	// along in this same insert since it's a column on this same row - see
	// parentData["is_requested_for_engr"] in Quotation.cs. The client only sends the
	// flag, not a timestamp, so stamp it here the same way RequestQuotationForEngr
	// (the separate, already-saved-record endpoint) already does.
	if body.SalesQuotation.IsRequestedForEngr && body.SalesQuotation.RequestedForEngrDate == "" {
		body.SalesQuotation.RequestedForEngrDate = time.Now().Format("01/02/2006 3:04:05 PM")
	}

	if err := services.DbInsert(tx, &body.SalesQuotation); err != nil {
		fmt.Println(err)
		fmt.Println("ERR", body)
		return body, fiber.StatusInternalServerError, errors.New("failed creating sales quotation")
	}

	// The latest-versions list reads through LatestQuotationView, which
	// DbInsert does not invalidate (it only clears model:SalesQuotation*).
	if err := invalidateLatestQuotationCache(); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed invalidating latest quotation cache")
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
// form) with no backing field anywhere.
//
// Changed 2026-09-03 (user decision, spec updated to match): this used to require
// naming one specific engineer (a picker modal, engrId validated against
// models.User) and scope the Engineering Sales Quotation List to only that
// engineer via requested_engr_id - see GetEngineeringQuotationListByEngr, which no
// longer filters on it. Checking the request is now the whole action: it just
// flips is_requested_for_engr, and every engineer sees every checked quotation on
// a shared list (the view's own WHERE is_requested_for_engr = 1 is what actually
// gates visibility now). RequestedEngrId/RequestedEngrName stay on the model/table
// but are no longer populated - nothing names a target engineer anymore.
// Phase 4 item 4.1.
func RequestQuotationForEngr(tx *gorm.DB, quotationId uint, at models.At) (*models.SalesQuotation, int, error) {
	var quotation models.SalesQuotation
	if err := tx.First(&quotation, quotationId).Error; err != nil {
		return nil, fiber.StatusNotFound, errors.New("quotation not found")
	}

	quotation.IsRequestedForEngr = true
	quotation.RequestedForEngrDate = time.Now().Format("01/02/2006 3:04:05 PM")

	if err := services.DbUpdate(tx, &quotation, map[string]interface{}{"id": quotation.ID}); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed requesting quotation for engineering")
	}

	if err := invalidateLatestQuotationCache(); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed invalidating latest quotation cache")
	}

	// DbUpdate's own InvalidateCacheByModel only ever clears model:SalesQuotation*
	// keys - the Engineering Sales Quotation List reads a completely different
	// cached type (EngineeringRedboxQuotationListView, bound to
	// vw_get_engineering_redbox_quotation_list) that also depends on this same row.
	// Same gap already found and fixed for BPI Financing/CustomerView/
	// SupplierTradeView - without this, a freshly requested quotation stays
	// invisible in Engineering until the cache happens to expire on its own.
	if err := InvalidateEngineeringQuotationListCache(); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed invalidating engineering quotation list cache")
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

// CancelQuotationForEngr reverses RequestQuotationForEngr - added on user request
// so REQUEST FOR ENGR. is a toggle (the button relabels to CANCEL REQUEST once
// checked) rather than a one-way action. Clears is_requested_for_engr, which is
// the view's only visibility gate (WHERE is_requested_for_engr = 1), so the
// quotation drops off every engineer's Sales Quotation List immediately - there is
// no separate "remove access" step needed. requested_for_engr_date is cleared too
// so a later re-request stamps a fresh date rather than showing the old one.
func CancelQuotationForEngr(tx *gorm.DB, quotationId uint, at models.At) (*models.SalesQuotation, int, error) {
	var quotation models.SalesQuotation
	if err := tx.First(&quotation, quotationId).Error; err != nil {
		return nil, fiber.StatusNotFound, errors.New("quotation not found")
	}

	quotation.IsRequestedForEngr = false
	quotation.RequestedForEngrDate = ""

	if err := services.DbUpdate(tx, &quotation, map[string]interface{}{"id": quotation.ID}); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed cancelling the engineering request")
	}

	if err := invalidateLatestQuotationCache(); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed invalidating latest quotation cache")
	}

	if err := InvalidateEngineeringQuotationListCache(); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed invalidating engineering quotation list cache")
	}

	atdata := models.SalesQuotationAt{
		RefId:                 quotation.ID,
		SalesQuotationContent: quotation.SalesQuotationContent,
		At:                    at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	return &quotation, fiber.StatusOK, nil
}

// for finalizing the quotation
func UpdateQuotationQuick(tx *gorm.DB, Quotation models.SalesQuotation, at models.At, conditions map[string]interface{}) error {
	// VALID UNTIL and the quote's reservations move together (10.4.5). A reservation window
	// can restart on the server - a quantity change, or keeping a reservation on hold at its
	// limit - while this form still holds the VALID UNTIL it loaded, and it sends the whole
	// header back. So an incoming VALID UNTIL counts as an edit only when DATE or VALIDITY
	// (DAYS), the two it is computed from on screen (5.1), changed with it; otherwise the
	// stored value stands.
	var stored models.SalesQuotation
	if Quotation.ID != 0 {
		if err := tx.Select("id", "date", "validity_days", "valid_until").Where("id = ?", Quotation.ID).Limit(1).Find(&stored).Error; err != nil {
			return errors.New("failed reading quotation")
		}
	}
	validityEdited := stored.ID == 0 ||
		!sameQuoteMoment(stored.Date, Quotation.Date) ||
		strings.TrimSpace(stored.ValidityDays) != strings.TrimSpace(Quotation.ValidityDays)
	if !validityEdited && stored.ValidUntil != "" {
		Quotation.ValidUntil = stored.ValidUntil
	}

	if err := services.DbUpdate(tx, &Quotation, conditions); err != nil {
		return errors.New("failed updating quotation")
	}

	if err := invalidateLatestQuotationCache(); err != nil {
		return errors.New("failed invalidating latest quotation cache")
	}

	if validityEdited {
		if err := moveQuoteReservationLimits(tx, Quotation.ID, Quotation.ValidUntil); err != nil {
			return errors.New("failed moving the quote's reservation limits")
		}
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
