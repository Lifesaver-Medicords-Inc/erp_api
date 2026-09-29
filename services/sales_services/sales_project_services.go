package sales_services

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
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

	// The item sets (tabs) the user removed on purpose - Remove Tabs in Sales. An item set is
	// deleted ONLY if its id is here. The diff alone used to be enough, and the diff reports
	// "removed" for any tab a payload merely fails to mention: one incomplete save deleted
	// every tab on Q#0026 five times over. Engineering never sends this, and cannot remove a
	// tab at all.
	RemovedItemSetIds []uint `json:"removed_item_set_ids"`
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

// writeChangedFields writes exactly the fields a save reports as changed - the keys of its
// Changes - with each value taken, correctly typed, from the row the client sent. Nothing
// else on the row is touched.
//
// Sales and Engineering edit the same project quotation at the same time (live sync,
// 2026-09-24). Writing the WHOLE row the client sent, as this used to (DbUpdate), meant a save
// carried the saver's copy of every field - including one the other app had changed a moment
// before and this screen had not caught up with yet. Sales editing FLOW while Engineering
// edited HEAD on the same content row: whichever saved second put the other's value back
// (user-reported 2026-09-26: "when update the advance conditions and items using sales and
// engineering it will not update"). Each client's change lists name every stored field of
// its section, so nothing a save means to change is lost by writing only what it lists.
//
// A change whose NewValue is null was not sent by that client (Engineering omits what it may
// not edit; nullable dropdown ids are only sent when set) and is left alone. A deliberate
// zero, blank or false IS written: GORM's UpdateColumns(struct) skips zero values, so "-- No
// Template --", a dropdown back to "-- Select --", an untick or a cleared field used to look
// saved and come back unchanged (2026-09-24) - a column map writes them. Keys that are not
// columns (size_up, final_selection - child lists) are skipped; the caller syncs those.
func writeChangedFields(tx *gorm.DB, row interface{}, keyColumn string, keyValue interface{}, changes map[string]FieldChange[any]) error {
	if len(changes) == 0 {
		return nil
	}

	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(row); err != nil {
		return err
	}

	rv := reflect.ValueOf(row)
	for rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	cols := map[string]interface{}{}
	for key, change := range changes {
		if change.NewValue == nil {
			continue
		}
		field := stmt.Schema.LookUpField(key)
		if field == nil || field.DBName == "" || field.PrimaryKey {
			continue
		}
		value, _ := field.ValueOf(tx.Statement.Context, rv)
		cols[field.DBName] = value
	}

	if len(cols) == 0 {
		return nil
	}

	if err := tx.Model(row).Where(keyColumn+" = ?", keyValue).UpdateColumns(cols).Error; err != nil {
		return err
	}
	return services.InvalidateCacheByModel(row)
}

// changed reports whether a save's Changes name this field.
func changed(changes map[string]FieldChange[any], key string) bool {
	_, ok := changes[key]
	return ok
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
	// Was missing entirely, so REQUEST FOR ENGR. could never be changed on an
	// existing project quote: the client's diff is unmarshalled into this struct, and
	// a field with no home here is discarded before applyQuotationFieldChanges ever
	// runs. The flag worked only on the INSERT that creates a quotation (§5.1, §6.3).
	IsRequestedForEngr   *FieldChange[any] `json:"is_requested_for_engr,omitempty"`
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

// Phase 2 - paged project headers + lazy project detail. Same split as the
// quick side: the module list pages latest headers only, and opening a project
// pulls its full tab/item/wiring tree on demand.
const projectOnlyWhere = "(project_name IS NOT NULL AND project_name <> '')"

const projectHeaderPageSize = 50

// SalesProjectDetail is the GetSalesProjects wire shape, named so the
// single-project detail endpoint can return it too.
type SalesProjectDetail struct {
	SalesQuotation                       []models.SalesQuotation                 `json:"SalesQuotation"`
	SalesProjectMultiplier               []models.SalesProjectMultiplier         `json:"sales_project_multiplier"`
	SalesProjectHistory                  []models.SalesProjectHistory            `json:"sales_project_history"`
	SalesProjectItemSet                  []models.SalesProjectItemSet            `json:"sales_project_item_set"`
	SalesProjectContent                  []models.SalesProjectContent            `json:"sales_project_content"`
	SalesProjectContentAdvancedCondition []models.SalesProjectAdvancedConditions `json:"sales_project_content_advanced_condition"`
	SalesProjectItems                    []models.SalesProjectItems              `json:"sales_project_items"`
	SalesProjectWirings                  []models.SalesProjectWiring             `json:"sales_project_wiring"`
	SalesProjectItemsSelectedImages      []models.SalesQuotationSelectedImage    `json:"sales_project_items_selected_images"`

	// The names behind the header's lookup ids, for a screen that shows the header without
	// the Setup lists Sales binds its dropdowns to - the Engineering app, which shows the
	// whole header read-only (user decision 2026-09-28). Added alongside; nothing else reads it.
	HeaderNames SalesProjectHeaderNames `json:"header_names"`
}

type SalesProjectHeaderNames struct {
	Application  string `json:"application"`
	PaymentTerms string `json:"payment_terms"`
	ShipType     string `json:"ship_type"`
	ShipTo       string `json:"ship_to"`
	BillTo       string `json:"bill_to"`
}

func newSalesProjectDetail() SalesProjectDetail {
	return SalesProjectDetail{
		SalesQuotation:                       []models.SalesQuotation{},
		SalesProjectMultiplier:               []models.SalesProjectMultiplier{},
		SalesProjectHistory:                  []models.SalesProjectHistory{},
		SalesProjectItemSet:                  []models.SalesProjectItemSet{},
		SalesProjectContent:                  []models.SalesProjectContent{},
		SalesProjectContentAdvancedCondition: []models.SalesProjectAdvancedConditions{},
		SalesProjectItems:                    []models.SalesProjectItems{},
		SalesProjectWirings:                  []models.SalesProjectWiring{},
		SalesProjectItemsSelectedImages:      []models.SalesQuotationSelectedImage{},
	}
}

// GetProjectHeadersPaged serves the project list: one page of latest project
// headers, newest first, no children. Same keyset contract as
// GetQuotationHeadersPaged (after/before/at, all unset = first page).
// Deliberately uncached, like GetItemsPaged.
func GetProjectHeadersPaged(after, before, at int) (quotationListShape, utils.PaginationMeta, int, error) {
	shape := newQuotationListShape()
	pagination := utils.PaginationMeta{PageSize: projectHeaderPageSize}

	var latest []models.LatestQuotationView
	query := initializers.DB.Model(&models.LatestQuotationView{}).Where(projectOnlyWhere)
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

	if err := query.Limit(projectHeaderPageSize).Find(&latest).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed getting project headers")
	}

	if reversed {
		for i, j := 0, len(latest)-1; i < j; i, j = i+1, j-1 {
			latest[i], latest[j] = latest[j], latest[i]
		}
	}

	var total int64
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(projectOnlyWhere).Count(&total).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed counting project headers")
	}
	pagination.Total = total
	pagination.TotalPages = int((total + int64(projectHeaderPageSize) - 1) / int64(projectHeaderPageSize))

	if len(latest) == 0 {
		return shape, pagination, 0, nil
	}

	var newer, older int64
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(projectOnlyWhere).Where("id > ?", latest[0].ID).Count(&newer).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed checking for a newer header page")
	}
	if err := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(projectOnlyWhere).Where("id < ?", latest[len(latest)-1].ID).Count(&older).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed checking for an older header page")
	}
	pagination.HasPrev = newer > 0
	pagination.HasNext = older > 0
	pagination.Page = int(newer)/projectHeaderPageSize + 1

	for _, v := range latest {
		shape.SalesQuotation = append(shape.SalesQuotation, v.SalesQuotation)
	}

	return shape, pagination, 0, nil
}

// SearchProjects backs the project Search: flat matching project headers,
// 20 to a page, no children - the picked record reopens through the headers'
// `at`. A blank term matches nothing (never the whole table).
func SearchProjects(term string, page int) (quotationListShape, utils.PaginationMeta, int, error) {
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

	// The latest copy of each document only - the same rows << PREV / NEXT >> walk. Searching
	// the table itself listed every copy, so each EDIT (a new sub-version since 2026-09-25)
	// showed up as another project in the list (user-reported 2026-09-26: "it not suppose to
	// create a new project if edit -> save happen"). Older copies are reached through the
	// version table (double-click the version number).
	base := initializers.DB.Model(&models.LatestQuotationView{}).
		Where(projectOnlyWhere).
		Where("document_no LIKE ?", "%"+term+"%")

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed searching projects")
	}
	pagination.Total = total
	pagination.TotalPages = int((total + int64(searchPageSize) - 1) / int64(searchPageSize))
	pagination.HasPrev = page > 1
	pagination.HasNext = page < pagination.TotalPages

	var latest []models.LatestQuotationView
	if err := base.Order("id DESC").
		Offset((page - 1) * searchPageSize).
		Limit(searchPageSize).
		Find(&latest).Error; err != nil {
		return shape, pagination, fiber.StatusInternalServerError, errors.New("failed searching projects")
	}
	for _, v := range latest {
		shape.SalesQuotation = append(shape.SalesQuotation, v.SalesQuotation)
	}

	return shape, pagination, 0, nil
}

// GetProjectVersions returns every version of ONE project document with the
// same scoping the quick versions endpoint uses, for the project
// version-history navigation that used to need the whole project list.
func GetProjectVersions(documentNo string) (quotationListShape, int, error) {
	shape := newQuotationListShape()

	bare := stripQuotationPrefix(documentNo)
	if bare == "" {
		return shape, fiber.StatusBadRequest, errors.New("missing document number")
	}

	var all []models.SalesQuotation
	if err := services.DbGet(&all, nil); err != nil {
		return shape, fiber.StatusInternalServerError, errors.New("failed getting project versions")
	}

	ids := make(map[uint]bool)
	for _, q := range all {
		if q.ProjectName != "" && stripQuotationPrefix(q.DocumentNo) == bare {
			shape.SalesQuotation = append(shape.SalesQuotation, q)
			ids[q.ID] = true
		}
	}

	if len(shape.SalesQuotation) == 0 {
		return shape, fiber.StatusNotFound, errors.New("project not found")
	}

	_ = ids
	return shape, 0, nil
}

// GetProjectDetail returns ONE project's full tree (tabs, contents, items,
// wiring, history, multipliers, images) scoped to that project id, for lazy
// loading when a project is opened. 404 when the id is unknown or not a
// project. Uncached: a handful of indexed seeks, like the item page children.
func GetProjectDetail(id int) (SalesProjectDetail, int, error) {
	detail := newSalesProjectDetail()

	if id <= 0 {
		return detail, fiber.StatusBadRequest, errors.New("missing or invalid project id")
	}

	var header models.SalesQuotation
	if err := services.DbGetNoCache(&header, map[string]interface{}{"id": id}); err != nil {
		return detail, fiber.StatusInternalServerError, errors.New("failed getting project")
	}
	if header.ID == 0 || header.ProjectName == "" {
		return detail, fiber.StatusNotFound, errors.New("project not found")
	}
	detail.SalesQuotation = append(detail.SalesQuotation, header)

	// Best effort: a missing name shows blank, it does not fail the quotation.
	initializers.DB.Raw(`
		SELECT ISNULL(a.name, '') AS application, ISNULL(p.name, '') AS payment_terms,
		       ISNULL(s.ship_name, '') AS ship_type, ISNULL(st.location, '') AS ship_to,
		       ISNULL(bt.location, '') AS bill_to
		  FROM tbl_trans_sales_quotation q
		  LEFT JOIN tbl_setup_application a    ON a.id = q.application_id
		  LEFT JOIN tbl_setup_payment_terms p  ON p.id = q.payment_terms_id
		  LEFT JOIN tbl_setup_ship_type s      ON s.id = q.ship_type_id
		  LEFT JOIN tbl_bpi_address st         ON st.id = q.ship_to_id
		  LEFT JOIN tbl_bpi_address bt         ON bt.id = q.bill_to_id
		 WHERE q.id = ?`, id).Scan(&detail.HeaderNames)

	uid := uint(id)
	byProject := map[string]interface{}{"based_id": uid}

	if err := services.DbGetNoCache(&detail.SalesProjectMultiplier, byProject); err != nil {
		return detail, fiber.StatusInternalServerError, errors.New("failed getting project multipliers")
	}

	if err := services.DbGetNoCache(&detail.SalesProjectItemSet, byProject); err != nil {
		return detail, fiber.StatusInternalServerError, errors.New("failed getting project item sets")
	}

	itemSetIDs := make([]uint, 0, len(detail.SalesProjectItemSet))
	for _, s := range detail.SalesProjectItemSet {
		itemSetIDs = append(itemSetIDs, s.ItemSetID)
	}

	// History rows hang off the project AND its item sets (see
	// GetFullProjectHistory on the client).
	historyIDs := append([]uint{uid}, itemSetIDs...)
	if err := services.DbGetNoCache(&detail.SalesProjectHistory, map[string]interface{}{"based_id": historyIDs}); err != nil {
		return detail, fiber.StatusInternalServerError, errors.New("failed getting project history")
	}

	if len(itemSetIDs) > 0 {
		byItemSet := map[string]interface{}{"based_id": itemSetIDs}
		// WITH Size Up and Final Selection. A plain read left both off every content row,
		// so every project the Sales app opened - and every live refresh - showed those two
		// grids empty, and the next save of that tab's content sent the empty lists back,
		// which syncContentChildren reads as "delete them all". The full /sales/projects list
		// always preloaded them; this per-project read did not.
		if err := services.DbGetWithPreloadsNoCache(&detail.SalesProjectContent, byItemSet,
			"SalesProjectContentFinal", "SalesProjectSizeUp"); err != nil {
			return detail, fiber.StatusInternalServerError, errors.New("failed getting project content")
		}
		if err := services.DbGetNoCache(&detail.SalesProjectContentAdvancedCondition, byItemSet); err != nil {
			return detail, fiber.StatusInternalServerError, errors.New("failed getting project conditions")
		}
		if err := services.DbGetNoCache(&detail.SalesProjectItems, byItemSet); err != nil {
			return detail, fiber.StatusInternalServerError, errors.New("failed getting project items")
		}
		if err := services.DbGetNoCache(&detail.SalesProjectWirings, byItemSet); err != nil {
			return detail, fiber.StatusInternalServerError, errors.New("failed getting project wiring")
		}

		itemIDs := make([]uint, 0, len(detail.SalesProjectItems))
		for _, it := range detail.SalesProjectItems {
			itemIDs = append(itemIDs, it.ItemsID)
		}
		if len(itemIDs) > 0 {
			if err := services.DbGetNoCache(&detail.SalesProjectItemsSelectedImages,
				map[string]interface{}{"quotation_quick_id": itemIDs}); err != nil {
				return detail, fiber.StatusInternalServerError, errors.New("failed getting project images")
			}
		}
	}

	return detail, 0, nil
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
	// An Edit session's copy is not built from a posted screen: StartProjectSubVersion copies
	// the stored project, every column of every row (see project_sub_version.go).
	if strings.EqualFold(strings.TrimSpace(body.VersionAction), "new_sub_version") {
		return body, fiber.StatusBadRequest, errors.New("a new sub-version of a project is started with POST /sales/projects/:id/sub-version")
	}

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

	// Project quotes share the header table behind vw_latest_sales_quotation,
	// so a new project also changes the latest-versions list.
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

	// Only the newest copy of a version can change. An Edit session makes a new sub-version,
	// and the one it came from becomes history: a late save from someone still looking at
	// the old copy (Engineering, before it has followed; a second window) must not land
	// there, where nobody would ever see it again. And an FQ never changes at all (§5.2).
	if status, err := guardProjectCopyIsCurrent(tx, body.ID); err != nil {
		return body, status, err
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
	removalRequested := map[uint]bool{}
	for _, id := range body.RemovedItemSetIds {
		removalRequested[id] = true
	}

	for _, tab := range body.Tabs {
		// A tab whose item set this save would delete without being asked to is skipped
		// whole - not just the deletion. Its diff also lists every one of its items and
		// wiring rows as removed, and applying those would empty the tab just the same.
		unrequested := false
		for _, removed := range tab.SalesProjectItemSet.Removed {
			if !removalRequested[removed.ItemSetID] {
				unrequested = true
				break
			}
		}
		if unrequested {
			fmt.Printf("UpdateSalesProject %d: refused to delete item set %d - its removal was not requested\n", body.ID, tab.BasedId)
			continue
		}

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

// deleteItemSetChildren removes every row that belongs to one item set, children before
// parents (Size Up and Final Selection carry foreign keys to their content row). Items give
// back any stock they had reserved (10.4.3). Change History is kept: it is the audit trail,
// and the removal itself is logged as a PROJECT line by the client.
//
// Safe to run when the same save's diff also lists the tab's items and wiring as removed:
// whichever runs second finds nothing left to delete.
func deleteItemSetChildren(tx *gorm.DB, itemSetID uint, at models.At) error {
	var contents []models.SalesProjectContent
	if err := tx.Where("based_id = ?", itemSetID).Find(&contents).Error; err != nil {
		return fmt.Errorf("item set remove (content read): %w", err)
	}
	for _, content := range contents {
		var sizeUps []models.SalesProjectSizeUp
		if err := tx.Where("sales_project_content_id = ?", content.ContentID).Find(&sizeUps).Error; err != nil {
			return fmt.Errorf("item set remove (size up read): %w", err)
		}
		for _, row := range sizeUps {
			if err := DeleteProjectSizeUp(tx, row, at); err != nil {
				return fmt.Errorf("item set remove (size up): %w", err)
			}
		}

		var finals []models.SalesProjectContentFinal
		if err := tx.Where("sales_project_content_id = ?", content.ContentID).Find(&finals).Error; err != nil {
			return fmt.Errorf("item set remove (final read): %w", err)
		}
		for _, row := range finals {
			if err := DeleteProjectContentFinal(tx, row, at); err != nil {
				return fmt.Errorf("item set remove (final): %w", err)
			}
		}

		if err := services.DbDelete(tx, &models.SalesProjectContent{}, map[string]interface{}{
			"content_id": content.ContentID,
		}); err != nil {
			return fmt.Errorf("item set remove (content): %w", err)
		}
	}

	if err := services.DbDelete(tx, &models.SalesProjectAdvancedConditions{}, map[string]interface{}{
		"based_id": itemSetID,
	}); err != nil {
		return fmt.Errorf("item set remove (advanced conditions): %w", err)
	}

	var items []models.SalesProjectItems
	if err := tx.Where("based_id = ?", itemSetID).Find(&items).Error; err != nil {
		return fmt.Errorf("item set remove (items read): %w", err)
	}
	for _, item := range items {
		if err := releaseProjectItemReservation(tx, item.ItemsID, at.AtUser); err != nil {
			return fmt.Errorf("item set remove (reservation release): %w", err)
		}
		if err := services.DbDelete(tx, &models.SalesProjectItems{}, map[string]interface{}{
			"items_id": item.ItemsID,
		}); err != nil {
			return fmt.Errorf("item set remove (item): %w", err)
		}
	}

	if err := services.DbDelete(tx, &models.SalesProjectWiring{}, map[string]interface{}{
		"based_id": itemSetID,
	}); err != nil {
		return fmt.Errorf("item set remove (wiring): %w", err)
	}

	return nil
}

// guardProjectCopyIsCurrent refuses a write to a finalized quote, or to a sub-version that a
// newer sub-version of the same version has replaced. The error names the current copy, so
// whoever gets it knows where the work now lives.
func guardProjectCopyIsCurrent(tx *gorm.DB, id uint) (int, error) {
	var current models.SalesQuotation
	if err := tx.First(&current, id).Error; err != nil {
		return fiber.StatusNotFound, errors.New("quotation not found")
	}
	code := models.FormatQuoteCode(current.DocumentNo, current.VersionNo, current.SubVersionNo)

	if current.IsFinalized {
		return fiber.StatusForbidden, fmt.Errorf("%s is finalized and cannot be changed", code)
	}

	var newer models.SalesQuotation
	err := tx.Where("document_no = ? AND version_no = ? AND TRY_CONVERT(int, sub_version_no) > TRY_CONVERT(int, ?)",
		current.DocumentNo, current.VersionNo, current.SubVersionNo).
		Order("TRY_CONVERT(int, sub_version_no) DESC").
		Limit(1).Find(&newer).Error
	if err != nil {
		return fiber.StatusInternalServerError, errors.New("failed checking for a newer copy of this quotation")
	}
	if newer.ID != 0 {
		return fiber.StatusConflict, fmt.Errorf("%s has been replaced by %s (id %d) - reopen the quotation to continue",
			code, models.FormatQuoteCode(newer.DocumentNo, newer.VersionNo, newer.SubVersionNo), newer.ID)
	}
	return 0, nil
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
	if fields.IsRequestedForEngr != nil {
		requested := fields.IsRequestedForEngr.NewValue
		updates["is_requested_for_engr"] = requested

		// The client sends the flag only, never a timestamp - same as the insert path
		// in quotation_service.go, which stamps it there for exactly this reason.
		// Turning the request off clears the date so a later re-request gets a fresh
		// one rather than reviving the original request's timestamp.
		if b, ok := requested.(bool); ok && b {
			updates["requested_for_engr_date"] = time.Now().Format("01/02/2006 3:04:05 PM")
		} else {
			updates["requested_for_engr_date"] = ""
		}
	}

	if len(updates) == 0 {
		return nil
	}

	if err := tx.Model(&models.SalesQuotation{}).
		Where("id = ?", id).
		Updates(updates).Error; err != nil {
		return err
	}

	// The quote's reservations follow its VALID UNTIL (10.4.5).
	if fields.ValidUntil != nil {
		if err := moveQuoteReservationLimits(tx, id, fmt.Sprint(fields.ValidUntil.NewValue)); err != nil {
			return err
		}
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
	// Same header table feeds vw_latest_sales_quotation - see
	// invalidateLatestQuotationCache in quotation_service.go.
	if err := invalidateLatestQuotationCache(); err != nil {
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
		// Everything filed under the tab goes with it. Deleting only the item set row left its
		// content, Size Up, Final Selection and Advanced Conditions behind as orphans (five of
		// each on the rehearsal database) - invisible, but still counted by anything that
		// reads those tables whole.
		if err := deleteItemSetChildren(tx, item.ItemSetID, at); err != nil {
			return 0, err
		}
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
// syncFinals / syncSizeUps: which of the two lists to make match what was sent. A list not
// synced is left exactly as stored.
//
// knownFinalIDs / knownSizeUpIDs: the row ids the saving screen had read. A stored row that
// was not sent is deleted only when it is named there - the screen saw it and took it off.
// One it never saw was added by the other app after the screen last read the list, and
// deleting it is how one app's pick wiped the other's. nil (a client that does not send
// them) keeps the old rule: the stored list becomes exactly what was sent.
//
// Fixed 2026-09-28 (user-reported: "size up and finals don't have websocket when
// updating ... final it no data sometimes"): the stored rows were read for the prune pass
// AFTER the new rows had been created, and a created row's id was never marked kept - so
// every pump picked on a saved quote was inserted and deleted again in the same save. The
// save reported success, the screen that picked it still showed it, the other app was told
// and found nothing new, and a reload showed nothing. The prune now runs over the rows as
// they were before this save, and a created row counts as kept.
func syncContentChildren(tx *gorm.DB, contentID uint, finals []models.SalesProjectContentFinal, sizeUps []models.SalesProjectSizeUp,
	syncFinals, syncSizeUps bool, knownFinalIDs, knownSizeUpIDs []uint, at models.At) error {
	// Resolve item ids BEFORE deduping, so the dedupe below can key on the pump
	// itself rather than falling back to comparing model text - two rows naming the
	// same pump are then caught even if one of them arrived without an id.
	resolveChildItemIds(tx, finals, sizeUps)

	finals = dedupeFinals(finals)
	sizeUps = dedupeSizeUps(sizeUps)

	if syncFinals {
		var existingFinals []models.SalesProjectContentFinal
		if err := tx.Where("sales_project_content_id = ?", contentID).Find(&existingFinals).Error; err != nil {
			return fmt.Errorf("content final read: %w", err)
		}

		keptFinals := map[uint]bool{}
		for _, v := range finals {
			// Sent without an id but naming a pump already on this list: that row, not a
			// second copy of it. A screen that has not yet been told the id its own pick was
			// given sends it like this.
			if v.ID == 0 {
				v.ID = matchStoredPump(len(existingFinals), func(i int) (uint, uint, string) {
					return existingFinals[i].ID, existingFinals[i].ItemID, existingFinals[i].Final
				}, v.ItemID, v.Final, keptFinals)
			}
			if v.ID > 0 {
				keptFinals[v.ID] = true
				if err := UpdateProjectContentFinal(tx, v, at, map[string]interface{}{"id": v.ID}); err != nil {
					return fmt.Errorf("content final update: %w", err)
				}
				continue
			}
			id, err := createProjectContentFinal(tx, contentID, v, at)
			if err != nil {
				return fmt.Errorf("content final add: %w", err)
			}
			keptFinals[id] = true
		}

		known := idSet(knownFinalIDs)
		for _, row := range existingFinals {
			if keptFinals[row.ID] || (known != nil && !known[row.ID]) {
				continue
			}
			if err := DeleteProjectContentFinal(tx, row, at); err != nil {
				return fmt.Errorf("content final delete: %w", err)
			}
		}
	}

	if syncSizeUps {
		var existing []models.SalesProjectSizeUp
		if err := tx.Where("sales_project_content_id = ?", contentID).Find(&existing).Error; err != nil {
			return fmt.Errorf("size up read: %w", err)
		}

		kept := map[uint]bool{}
		for _, v := range sizeUps {
			if v.ID == 0 {
				v.ID = matchStoredPump(len(existing), func(i int) (uint, uint, string) {
					return existing[i].ID, existing[i].ItemID, existing[i].Model
				}, v.ItemID, v.Model, kept)
			}
			if v.ID > 0 {
				kept[v.ID] = true
				if err := UpdateProjectSizeUp(tx, v, at, map[string]interface{}{"id": v.ID}); err != nil {
					return fmt.Errorf("size up update: %w", err)
				}
				continue
			}
			id, err := createProjectSizeUp(tx, contentID, v, at)
			if err != nil {
				return fmt.Errorf("size up add: %w", err)
			}
			kept[id] = true
		}

		known := idSet(knownSizeUpIDs)
		for _, row := range existing {
			if kept[row.ID] || (known != nil && !known[row.ID]) {
				continue
			}
			if err := DeleteProjectSizeUp(tx, row, at); err != nil {
				return fmt.Errorf("size up delete: %w", err)
			}
		}
	}

	return nil
}

// matchStoredPump finds the stored row a row sent without an id stands for: the same pump by
// item id when both have one, otherwise the same model name. Rows already claimed by this
// save are skipped. 0 when there is none.
func matchStoredPump(n int, stored func(i int) (id uint, itemID uint, model string), itemID uint, model string, claimed map[uint]bool) uint {
	model = strings.TrimSpace(model)
	for i := 0; i < n; i++ {
		id, storedItem, storedModel := stored(i)
		if claimed[id] {
			continue
		}
		if itemID > 0 && storedItem > 0 {
			if itemID == storedItem {
				return id
			}
			continue
		}
		if model != "" && strings.EqualFold(strings.TrimSpace(storedModel), model) {
			return id
		}
	}
	return 0
}

// idSet is nil for a nil list - "not sent" - and an empty set for an empty one.
func idSet(ids []uint) map[uint]bool {
	if ids == nil {
		return nil
	}
	set := make(map[uint]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
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

		if err := syncContentChildren(tx, item.ContentID, addFinals, addSizeUps, true, true, nil, nil, at); err != nil {
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

		if err := writeChangedFields(tx, &entry.Item, "content_id", entry.Item.ContentID, entry.Changes); err != nil {
			return fmt.Errorf("content update: %w", err)
		}

		// DbUpdate does not touch associations, so without this the finals and size-ups
		// on an edited tab were discarded on every save.
		// Only the list this save changed. Syncing both replaced the other one with whatever
		// this client last read - a Sales edit of FLOW wiped the Size Up an engineer had just
		// picked.
		if err := syncContentChildren(tx, entry.Item.ContentID, entry.Item.SalesProjectContentFinal, entry.Item.SalesProjectSizeUp,
			changed(entry.Changes, "final_selection"), changed(entry.Changes, "size_up"),
			entry.Item.FinalKnownIDs, entry.Item.SizeUpKnownIDs, at); err != nil {
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
		if err := writeChangedFields(tx, &entry.Item, "conditions_id", entry.Item.ConditionsID, entry.Changes); err != nil {
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
		// A removed item's reservation goes with it; anything it held returns (10.4.3).
		if err := releaseProjectItemReservation(tx, item.ItemsID, at.AtUser); err != nil {
			return fmt.Errorf("item reservation release: %w", err)
		}
	}
	for _, entry := range diff.Updated {
		if err := writeChangedFields(tx, &entry.Item, "items_id", entry.Item.ItemsID, entry.Changes); err != nil {
			return fmt.Errorf("item update: %w", err)
		}
		// An edited QTY is a new reservation request with a fresh window (10.4.5). Skipped at
		// zero: DbUpdate ignores zero values too, so a zero here means QTY was not sent.
		// Only when this save changed QTY - otherwise Qty is this client's copy, which may be
		// older than the stored one.
		if changed(entry.Changes, "qty") && entry.Item.Qty > 0 {
			if err := syncProjectItemReservation(tx, entry.Item.ItemsID, entry.Item.Qty); err != nil {
				return fmt.Errorf("item reservation sync: %w", err)
			}
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
		if err := writeChangedFields(tx, &entry.Item, "id", entry.Item.ID, entry.Changes); err != nil {
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
