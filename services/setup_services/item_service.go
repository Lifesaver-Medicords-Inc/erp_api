package setup_services

import (
	"errors"
	"fmt"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/models/inventory_models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
	"gorm.io/gorm"
)

type Body struct {
	models.Item
	ItemSpecs       models.ItemSpecs       `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecs `json:"additional_specs"`
}

type SaveBody struct {
	models.Item
	TradeTypeId     []uint                       `json:"trade_type_id"`
	ItemSpecs       models.ItemSpecs             `json:"itemspecs"`
	AdditionalSpecs models.AdditionalSpecsSchema `json:"additionalspecs"`
	ItemImages      ItemImage                    `json:"itemimages"`
	ItemInventory   models.ItemInventory         `json:"iteminventory"`
}
type ItemSpecsWithTemplate struct {
	ItemSpecs      models.ItemSpecs `json:"itemspecs"`
	TemplateFields []Field          `json:"template_fields"`
}

type ItemImage struct {
	NewImages     []models.ItemImage `json:"newimages"`
	ReplaceImages []models.ItemImage `json:"replaceimages"`
	DeleteImages  []models.ItemImage `json:"deleteimages"`
}

type ItemSpecs struct {
	Template           string       `json:"template"`
	Fields             []SpecsField `json:"fields"`
	ManufacturerOrigin string       `json:"manufacturer_origin"`
	Fla1               string       `json:"fla_1"`
	Fla2               string       `json:"fla_2"`
	Volt1              string       `json:"volt_1"`
	Volt2              string       `json:"volt_2"`
}

type SpecsField struct {
	Title string `json:"title"`
	Value string `json:"value"`
}

func GetItems(conditions map[string]interface{}) (interface{}, int, error) {
	type Response struct {
		Items            []models.ItemView                        `json:"items"`
		ItemSpecs        []models.ItemSpecs                       `json:"itemspecs"`
		AdditionalSpecs  []models.AdditionalSpecsView             `json:"additionalspecs"`
		ItemImage        []models.ItemImage                       `json:"itemimages"`
		ItemPurchasing   []models.ItemPurchasingView              `json:"itempurchasing"`
		ItemSales        []models.ItemSalesView                   `json:"itemsales"`
		ItemInventory    []models.ItemInventory                   `json:"iteminventory"`
		ItemAvailableInv []models.ItemAvailableInventoryModelView `json:"itemavailableinv"`
		ItemProductions  []models.ItemProductionView              `json:"itemproduction"`
	}

	var response Response
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	// Run all 9 queries in parallel
	queries := []struct {
		name string
		fn   func() error
	}{
		{"items", func() error { return services.DbGet(&response.Items, conditions) }},
		{"itemspecs", func() error {
			return services.DbGetWithPreloads(&response.ItemSpecs, conditions, "ItemSpecsTemplate")
		}},
		{"additionalspecs", func() error { return services.DbGet(&response.AdditionalSpecs, conditions) }},
		{"itemimage", func() error { return services.DbGet(&response.ItemImage, conditions) }},
		{"itempurchasing", func() error { return services.DbGet(&response.ItemPurchasing, conditions) }},
		{"itemsales", func() error { return services.DbGet(&response.ItemSales, conditions) }},
		{"iteminventory", func() error { return services.DbGet(&response.ItemInventory, conditions) }},
		{"itemavailableinv", func() error { return services.DbGet(&response.ItemAvailableInv, conditions) }},
		{"itemproductions", func() error { return services.DbGet(&response.ItemProductions, conditions) }},
	}

	errorMap := map[string]string{
		"items":            "failed getting items",
		"itemspecs":        "failed getting item spec",
		"additionalspecs":  "failed getting item additional spec",
		"itemimage":        "failed getting item image",
		"itempurchasing":   "failed getting item purchasing",
		"itemsales":        "failed getting item sales",
		"iteminventory":    "failed getting item production",
		"itemavailableinv": "failed getting item available inventory",
		"itemproductions":  "failed getting item production",
	}

	for _, q := range queries {
		wg.Add(1)
		go func(query struct {
			name string
			fn   func() error
		}) {
			defer wg.Done()
			if err := query.fn(); err != nil {
				mu.Lock()
				if firstErr == nil {
					firstErr = errors.New(errorMap[query.name])
				}
				mu.Unlock()
			}
		}(q)
	}

	wg.Wait()

	if firstErr != nil {
		return response, fiber.StatusInternalServerError, firstErr
	}

	return response, 0, nil
}

// Item Entry loads one page of items at a time. GetItems above still returns the whole
// catalogue for the callers that genuinely need every row - the Sales quotation/order item
// pickers, the Engineering pump picker, Item Model Setup and Item Stock Add.
const itemPageSize = 20

// The same nine lists GetItems returns, so the client binds one shape whichever route it calls.
type ItemPageResponse struct {
	Items            []models.ItemView                        `json:"items"`
	ItemSpecs        []models.ItemSpecs                       `json:"itemspecs"`
	AdditionalSpecs  []models.AdditionalSpecsView             `json:"additionalspecs"`
	ItemImage        []models.ItemImage                       `json:"itemimages"`
	ItemPurchasing   []models.ItemPurchasingView              `json:"itempurchasing"`
	ItemSales        []models.ItemSalesView                   `json:"itemsales"`
	ItemInventory    []models.ItemInventory                   `json:"iteminventory"`
	ItemAvailableInv []models.ItemAvailableInventoryModelView `json:"itemavailableinv"`
	ItemProductions  []models.ItemProductionView              `json:"itemproduction"`
}

// GetItemsPaged returns one page of items, oldest first, with only that page's child rows.
//
// Keyset rather than OFFSET: ids are indexed and never renumber, so a page's children come back
// on a 20-id IN list instead of one covering all 3,707 items - which is what pushed the
// ItemSpecsTemplate preload past SQL Server's 2,100-parameter ceiling on 2026-09-15 and left
// Item Entry blank.
//
//	after  - the last id of the page just shown (NEXT >>)
//	before - the first id of it (<< PREV)
//	at     - open the page starting at this item (a search hit, or one just saved)
//
// All three unset means the first page.
func GetItemsPaged(after, before, at int) (ItemPageResponse, utils.PaginationMeta, int, error) {
	var response ItemPageResponse

	query := initializers.DB.Model(&models.ItemView{})
	reversed := false

	switch {
	case before > 0:
		query = query.Where("id < ?", before).Order("id DESC")
		reversed = true
	case after > 0:
		query = query.Where("id > ?", after).Order("id ASC")
	case at > 0:
		query = query.Where("id >= ?", at).Order("id ASC")
	default:
		query = query.Order("id ASC")
	}

	if err := query.Limit(itemPageSize).Find(&response.Items).Error; err != nil {
		return response, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed getting items")
	}

	// A PREV page is read back to front; put it the right way round before anything else sees it.
	if reversed {
		for i, j := 0, len(response.Items)-1; i < j; i, j = i+1, j-1 {
			response.Items[i], response.Items[j] = response.Items[j], response.Items[i]
		}
	}

	var total int64
	if err := initializers.DB.Model(&models.Item{}).Count(&total).Error; err != nil {
		return response, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed counting items")
	}

	pagination := utils.PaginationMeta{PageSize: itemPageSize, Total: total}
	if len(response.Items) == 0 {
		return response, pagination, 0, nil
	}

	// Asked outright rather than inferred from fetching one extra row: the answer has to hold
	// travelling in both directions, and an indexed count over a few thousand ids is free.
	// The count of items BEHIND this page also gives the page number for free.
	countSide := func(where string, id uint) (int64, error) {
		var count int64
		err := initializers.DB.Model(&models.Item{}).Where(where, id).Count(&count).Error
		return count, err
	}

	itemsBefore, err := countSide("id < ?", response.Items[0].ID)
	if err != nil {
		return response, pagination, fiber.StatusInternalServerError, errors.New("failed checking for an earlier item page")
	}
	itemsAfter, err := countSide("id > ?", response.Items[len(response.Items)-1].ID)
	if err != nil {
		return response, pagination, fiber.StatusInternalServerError, errors.New("failed checking for a later item page")
	}

	pagination.HasPrev = itemsBefore > 0
	pagination.HasNext = itemsAfter > 0
	// Walking with after/before from the first page, this is exactly the page number. Opening
	// at an id lands mid-page by design (that item goes first), so it reads as the position of
	// that item rather than a boundary - which is the more useful answer on screen anyway.
	pagination.Page = int(itemsBefore)/itemPageSize + 1
	pagination.TotalPages = int((total + int64(itemPageSize) - 1) / int64(itemPageSize))

	ids := make([]uint, 0, len(response.Items))
	for _, item := range response.Items {
		ids = append(ids, item.ID)
	}

	basedIn := map[string]interface{}{"based_id": ids}
	itemIn := map[string]interface{}{"item_id": ids}

	// Deliberately uncached - see DbGetWithPreloadsNoCache. Eight of these are one query each
	// against an indexed foreign key over 20 ids.
	queries := []struct {
		name string
		fn   func() error
	}{
		{"item specs", func() error {
			return services.DbGetWithPreloadsNoCache(&response.ItemSpecs, basedIn, "ItemSpecsTemplate")
		}},
		{"item additional specs", func() error { return services.DbGetNoCache(&response.AdditionalSpecs, basedIn) }},
		{"item images", func() error { return services.DbGetNoCache(&response.ItemImage, basedIn) }},
		{"item purchasing", func() error { return services.DbGetNoCache(&response.ItemPurchasing, basedIn) }},
		{"item sales", func() error { return services.DbGetNoCache(&response.ItemSales, basedIn) }},
		{"item inventory", func() error { return services.DbGetNoCache(&response.ItemInventory, basedIn) }},
		{"item available inventory", func() error { return services.DbGetNoCache(&response.ItemAvailableInv, itemIn) }},
		{"item production", func() error { return services.DbGetNoCache(&response.ItemProductions, itemIn) }},
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	for _, q := range queries {
		wg.Add(1)
		go func(query struct {
			name string
			fn   func() error
		}) {
			defer wg.Done()
			if err := query.fn(); err != nil {
				mu.Lock()
				// Named, unlike GetItems' shared error map: when this page came up empty on
				// 2026-09-15, one message stood for nine possible causes and the real one
				// (the specs preload) had to be found by reading Redis keys.
				if firstErr == nil {
					firstErr = fmt.Errorf("failed getting %s: %w", query.name, err)
				}
				mu.Unlock()
			}
		}(q)
	}

	wg.Wait()

	if firstErr != nil {
		return response, pagination, fiber.StatusInternalServerError, firstErr
	}

	return response, pagination, 0, nil
}

// GetItemsSearch backs Item Entry's Search modal: a flat item list with no children, since the
// modal shows five columns. Numbered pages of itemPageSize with Previous/Next and a "Page x of
// y" count, matching the form behind it, so searching 3,707 items never draws more rows than
// the list it sits over. The page then reopens the chosen item through GetItemsPaged's `at`.
func GetItemsSearch(search string, page int) ([]models.ItemView, utils.PaginationMeta, int, error) {
	var items []models.ItemView

	if page < 1 {
		page = 1
	}

	searchColumns := []string{"item_code", "item_name", "item_model", "item_brand"}

	total, err := services.DbSearchPage(&items, nil, search, searchColumns, nil, page, itemPageSize, "id ASC")
	if err != nil {
		return items, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed searching items")
	}

	totalPages := int((total + int64(itemPageSize) - 1) / int64(itemPageSize))

	return items, utils.PaginationMeta{
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		PageSize:   itemPageSize,
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
	}, 0, nil
}

// ItemPickerRow is one row of the Sales quotation's item and model pickers.
//
// It carries exactly what those two modals display AND what SalesItemGridEditor.GetItemData
// writes onto a quotation line (unit_of_measure, item_model, item_name), so a picked item can
// be used without a second round trip - the caller merges this row into its local ItemList and
// every existing consumer keeps working unchanged.
//
// BomId is 0 when the item has no BOM. The modals show "BOM" or "SINGLE" from it, and the
// quotation branches on it (GetBomDataRecursive vs GetItemData), which is why it has to come
// from the server now that the client no longer holds the whole BOM table for this purpose.
type ItemPickerRow struct {
	ID            uint   `json:"id"`
	ItemCode      string `json:"item_code"`
	ItemName      string `json:"item_name"`
	ItemModel     string `json:"item_model"`
	ItemNameId    uint   `json:"item_name_id"`
	UnitOfMeasure string `json:"unit_of_measure"`
	BomId         uint   `json:"bom_id"`
}

// TOP 1, not a join: an item with two BOM heads would otherwise duplicate its own row in the
// picker. The client's old bomHead.Where(...).FirstOrDefault() had the same effect.
const itemPickerColumns = `v.id, v.item_code, v.item_name, v.item_model, v.item_name_id, v.unit_of_measure,
	ISNULL((SELECT TOP 1 b.id FROM tbl_setup_item_bom b WHERE b.item_id = v.id), 0) AS bom_id`

func itemPickerSearch(search string, args *[]interface{}) string {
	search = strings.TrimSpace(search)
	if search == "" {
		return ""
	}

	like := "%" + search + "%"
	*args = append(*args, like, like, like)
	return " AND (v.item_code LIKE ? OR v.item_name LIKE ? OR v.item_model LIKE ?)"
}

func runItemPickerQuery(where string, args []interface{}, page int) ([]ItemPickerRow, utils.PaginationMeta, error) {
	if page < 1 {
		page = 1
	}

	var total int64
	if err := initializers.DB.Raw("SELECT COUNT(*) FROM vw_items v WHERE "+where, args...).
		Scan(&total).Error; err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	// ORDER BY id, matching the order these pickers showed while they were fed the whole
	// catalogue in id order - a picker that suddenly re-sorts itself is a behaviour change.
	paged := append(append([]interface{}{}, args...), (page-1)*itemPageSize, itemPageSize)
	rows := []ItemPickerRow{}
	if err := initializers.DB.Raw(
		"SELECT "+itemPickerColumns+" FROM vw_items v WHERE "+where+
			" ORDER BY v.id ASC OFFSET ? ROWS FETCH NEXT ? ROWS ONLY", paged...).
		Scan(&rows).Error; err != nil {
		return nil, utils.PaginationMeta{}, err
	}

	totalPages := int((total + int64(itemPageSize) - 1) / int64(itemPageSize))

	return rows, utils.PaginationMeta{
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		PageSize:   itemPageSize,
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
	}, nil
}

// GetItemPickerNames backs the quotation's ITEM picker: one row per item_name_id, the server
// side of what SalesItemModal.distinctByItemNameId used to do in memory over every item. That
// dedup is a catalogue-wide GROUP BY, so it cannot be done on a 20-row page - which is exactly
// why this endpoint exists rather than the picker reusing /setup/item/search.
//
// MIN(id) picks the same representative row the client's GroupBy().First() did, since it read
// an id-ordered table.
func GetItemPickerNames(search string, page int) ([]ItemPickerRow, utils.PaginationMeta, int, error) {
	args := []interface{}{}
	where := "v.id IN (SELECT MIN(id) FROM vw_items GROUP BY item_name_id)" + itemPickerSearch(search, &args)

	rows, pagination, err := runItemPickerQuery(where, args, page)
	if err != nil {
		return nil, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed getting the item picker list")
	}

	return rows, pagination, 0, nil
}

// GetItemPickerModels backs the quotation's MODEL picker: every item sharing the given item's
// item_name_id - ModelModal's own filter, moved to the server. An item with no item_name_id
// (0) returns just itself, which is what the client's FirstOrDefault()==0 path did.
func GetItemPickerModels(itemID int, search string, page int) ([]ItemPickerRow, utils.PaginationMeta, int, error) {
	if itemID <= 0 {
		return nil, utils.PaginationMeta{}, fiber.StatusBadRequest, errors.New("item_id is required")
	}

	var nameID uint
	if err := initializers.DB.Raw("SELECT ISNULL(item_name_id, 0) FROM vw_items WHERE id = ?", itemID).
		Scan(&nameID).Error; err != nil {
		return nil, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed getting the item")
	}

	args := []interface{}{}
	where := "v.id = ?"
	if nameID != 0 {
		where = "v.item_name_id = ?"
		args = append(args, nameID)
	} else {
		args = append(args, itemID)
	}
	where += itemPickerSearch(search, &args)

	rows, pagination, err := runItemPickerQuery(where, args, page)
	if err != nil {
		return nil, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed getting the model list")
	}

	return rows, pagination, 0, nil
}

func GetItem(id int) (Body, int, error) {
	conditions := map[string]interface{}{
		"id": id,
	}
	var record Body

	if err := services.DbGet(&record.Item, conditions); err != nil {
		return record, fiber.StatusInternalServerError, errors.New("failed getting item")
	}

	itemSpecsConditions := map[string]interface{}{
		"based_id": record.ID,
	}

	if err := GetItemSpec(&record.ItemSpecs, itemSpecsConditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	if err := GetAdditionalSpec(&record.AdditionalSpecs, conditions); err != nil {
		return record, fiber.StatusInternalServerError, err
	}

	return record, 0, nil
}

func CreateItem(c *fiber.Ctx, tx *gorm.DB) (SaveBody, int, error) {
	var savebody SaveBody

	if err := c.BodyParser(&savebody); err != nil {
		return savebody, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	// Trello #091: ItemCode was computed client-side (frm_Item_Entry.cs took
	// the last-loaded item's code, +1, zero-padded) with no server-side
	// sequence behind it - two near-simultaneous creates, or a stale local
	// list, could compute and submit the same "next" code. Same class of bug
	// as generateCustomerCode/generateSupplierCode (bpi_entity_service.go) -
	// fixed the same way: the server generates the real code from the highest
	// number ever issued, ignoring whatever the client guessed and sent.
	generatedCode, err := generateItemCode(tx)
	if err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}
	savebody.ItemCode = generatedCode

	if err := services.DbInsert(tx, &savebody.Item); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating item")
		}

		return savebody, fiber.StatusInternalServerError, err
	}

	at := extractAtFromContext(c)

	atdata := models.ItemAt{
		RefId:       savebody.ID,
		ItemContent: savebody.ItemContent,
		At:          at}

	if err := services.DbInsert(tx, &atdata); err != nil {
		return savebody, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	for _, v := range savebody.TradeTypeId {
		if err := CreateItemTradeTypes(tx, savebody.ID, uint(v), at); err != nil {
			return savebody, 0, err
		}
	}

	if err := CreateItemSpec(tx, savebody.ID, savebody.ItemSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	if err := CreateAdditionalSpec(tx, savebody.ID, savebody.AdditionalSpecs, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	if err := CreateItemImageChild(tx, savebody.ID, savebody.ItemImages.NewImages, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}
	if err := CreateItemInventory(tx, savebody.ID, savebody.ItemInventory, at); err != nil {
		return savebody, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
	return savebody, 0, nil
}

func UpdateItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (SaveBody, int, error) {
	var body SaveBody
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if body.ID == 0 {
		return body, fiber.StatusBadRequest, errors.New("item id is required for update")
	}

	// Same guard as CreateItem, excluding this item's own row so re-saving an
	// item without changing its code doesn't reject against itself.
	if body.ItemCode != "" {
		var existingCount int64
		if err := tx.Model(&models.Item{}).
			Where("item_code = ? AND id <> ?", body.ItemCode, body.ID).
			Count(&existingCount).Error; err != nil {
			return body, fiber.StatusInternalServerError, errors.New("failed checking for an existing item with this code")
		}
		if existingCount > 0 {
			return body, fiber.StatusBadRequest, fmt.Errorf(
				"an item with code %s already exists", body.ItemCode)
		}
	}

	// Every field the Item Entry header edits, blanks and zeros included (see
	// services.DbUpdateFields). DbUpdate skipped zero values, so clearing a field, setting the
	// price to 0 or a dropdown back to none left the old value in place (user-reported
	// 2026-09-14). The item code is named only when one was sent, so it can never be blanked.
	itemFields := []string{"ItemNameId", "ItemClassId", "ItemBrandId", "UnitOfMeasureId", "ItemModel",
		"CatalogueYear", "ItemTangibilityType", "IsStopSelling", "Price"}
	if body.ItemCode != "" {
		itemFields = append(itemFields, "ItemCode")
	}
	if err := services.DbUpdateFields(tx, &body.Item, conditions, itemFields...); err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			err = errors.New("duplicate record error")
		} else {
			err = errors.New("failed creating item")
		}
		return body, fiber.StatusInternalServerError, err
	}

	at := extractAtFromContext(c)

	atdata := models.ItemAt{
		RefId:       body.ID,
		ItemContent: body.ItemContent,
		At:          at,
	}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed inserting itemat")
	}

	// ✅ Use body.ID explicitly — don't rely on it being set elsewhere
	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := UpdateTradeTypes(tx, body.ID, body.TradeTypeId, at); err != nil {
		return body, 0, err
	}

	if err := UpdateItemSpec(tx, body.ID, body.ItemSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateAdditionalSpec(tx, body.ID, body.AdditionalSpecs, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateItemImage(tx, body.ID, body.ItemImages, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	if err := UpdateItemInventory(tx, body.ID, body.ItemInventory, at, conditions); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()
	return body, 0, nil
}

// Trello #091: same fix as generateCustomerCode/generateSupplierCode
// (bpi_entity_service.go) - read the highest number ever issued (not a
// COUNT, and not "the last loaded row") so a gap or a concurrent create
// never collides with, or reuses, an already-issued code. item_code is
// stored as plain zero-padded digits with no prefix (the "I#" the UI shows
// is display-only, added in frm_Item_Entry.cs, never persisted).
func generateItemCode(tx *gorm.DB) (string, error) {
	var maxNum int
	if err := tx.Raw(`
		SELECT ISNULL(MAX(TRY_CAST(item_code AS INT)), 0)
		FROM tbl_setup_item
	`).Scan(&maxNum).Error; err != nil {
		return "", errors.New("failed computing next item_code")
	}

	return fmt.Sprintf("%04d", maxNum+1), nil
}

func DeleteItem(c *fiber.Ctx, tx *gorm.DB, conditions map[string]interface{}) (Body, int, error) {
	var body Body
	if err := c.BodyParser(&body); err != nil {
		return body, fiber.StatusBadRequest, errors.New("cannot bind request")
	}

	if err := services.DbDelete(tx, &body.Item, conditions); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed deleting item")
	}

	at := extractAtFromContext(c)

	atdata := models.ItemAt{RefId: body.ID, ItemContent: body.ItemContent, At: at}
	if err := services.DbInsert(tx, &atdata); err != nil {
		return body, fiber.StatusInternalServerError, errors.New("failed creating itemat")
	}

	conditions = map[string]interface{}{
		"based_id": body.ID,
	}

	if err := DeleteItemSpec(tx, conditions, at); err != nil {
		return body, fiber.StatusInternalServerError, err
	}

	InvalidateItemCaches()

	return body, 0, nil
}

func extractAtFromContext(c *fiber.Ctx) models.At {
	at, ok := c.Locals("at").(models.At)
	if !ok {
		at = models.At{}
	}
	return at
}

func InvalidateItemCaches() {
	cacheKeys := []interface{}{
		models.ItemView{},
		models.AdditionalSpecsView{},
		models.ItemImage{},
		models.ItemPurchasingView{},
		models.ItemSalesView{},
		models.ItemProductionView{},
		models.InvTrackerView{},
		models.PurchaseOrderView{},
		models.InvLogbookView{},
		models.AllBomView{},
		models.AllBinLocationView{},
		models.SalesOrderViewIR{},
		models.SalesOrderViewPA{},
		models.PurchaseOrderDetailsView{},
		models.ItemAvailableInventoryModelView{},
		accounting_models.ChartOfAccountViewList{},
		accounting_models.TaxView{},
		accounting_models.TaxDetailsView{},
		accounting_models.ChartClass{},
		accounting_models.ChartClassAt{},
		accounting_models.ChartOfAccounts{},
		accounting_models.ChartOfAccountsAt{},
		accounting_models.Tax{},
		accounting_models.TaxAt{},
		accounting_models.TaxDetails{},
		accounting_models.TaxDetailsAt{},
		accounting_models.SupplierTradeView{},
		accounting_models.PaymentVoucherDetailsGet{},
		inventory_models.WarehouseReceivingView{},
		inventory_models.PurchaseOrderDocView{},
		inventory_models.WarehouseReceivingAreaView{},
		inventory_models.PurchaseOrderReceivingView{},
		inventory_models.PurchaseOrderReceivingDetailsView{},
	}
	for _, key := range cacheKeys {
		services.InvalidateCache(services.GetKey(key, nil))
	}
}
