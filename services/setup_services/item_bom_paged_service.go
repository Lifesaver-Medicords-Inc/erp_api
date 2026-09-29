package setup_services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"github.com/pierceperado/smpc/utils"
)

// The BOM screen, 20 BOMs a page (user request 2026-09-26: "optimize of BOM 5,842 items and 130k
// for item details", "create pagination 20 item per page in BOM").
//
// GET /setup/bom returns every BOM head and every component line in the system - on test_fresh
// 5,842 heads and 130,380 lines - to show one BOM. These serve the screen a page at a time
// instead, with the same two lists in the same shape (bom_head, bom_details), so the screen's
// binding is unchanged; only its navigation learns about pages. GET /setup/bom is left as it
// is for its other callers.

const bomPageSize = 20

type BomPageResponse struct {
	ItemBom     []models.BomViewList     `json:"bom_head"`
	ItemDetails []models.BomViewItemList `json:"bom_details"`
}

// GetBomsPaged: one page of BOM heads in id order - the first page, the one after `after`, the
// one before `before`, or the one starting at `at` (a searched or just-saved BOM opens as the
// first row of its page) - with the component lines of those heads only. Uncached, like
// GetItemsPaged: a page is 20 ids and a few indexed seeks, and cached pages would go stale on
// every save.
func GetBomsPaged(after, before, at int) (BomPageResponse, utils.PaginationMeta, int, error) {
	response := BomPageResponse{ItemBom: []models.BomViewList{}, ItemDetails: []models.BomViewItemList{}}

	query := initializers.DB.Model(&models.BomViewList{})
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

	if err := query.Limit(bomPageSize).Find(&response.ItemBom).Error; err != nil {
		return response, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed getting bom page")
	}

	// A PREV page is read back to front; put it the right way round before anything else sees it.
	if reversed {
		for i, j := 0, len(response.ItemBom)-1; i < j; i, j = i+1, j-1 {
			response.ItemBom[i], response.ItemBom[j] = response.ItemBom[j], response.ItemBom[i]
		}
	}

	var total int64
	if err := initializers.DB.Model(&models.BomViewList{}).Count(&total).Error; err != nil {
		return response, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed counting boms")
	}

	pagination := utils.PaginationMeta{PageSize: bomPageSize, Total: total}
	if len(response.ItemBom) == 0 {
		return response, pagination, 0, nil
	}

	countSide := func(where string, id uint) (int64, error) {
		var count int64
		err := initializers.DB.Model(&models.BomViewList{}).Where(where, id).Count(&count).Error
		return count, err
	}

	bomsBefore, err := countSide("id < ?", response.ItemBom[0].ID)
	if err != nil {
		return response, pagination, fiber.StatusInternalServerError, errors.New("failed checking for an earlier bom page")
	}
	bomsAfter, err := countSide("id > ?", response.ItemBom[len(response.ItemBom)-1].ID)
	if err != nil {
		return response, pagination, fiber.StatusInternalServerError, errors.New("failed checking for a later bom page")
	}

	pagination.HasPrev = bomsBefore > 0
	pagination.HasNext = bomsAfter > 0
	pagination.Page = int(bomsBefore)/bomPageSize + 1
	pagination.TotalPages = int((total + int64(bomPageSize) - 1) / int64(bomPageSize))

	ids := make([]uint, 0, len(response.ItemBom))
	for _, head := range response.ItemBom {
		ids = append(ids, head.ID)
	}

	// Only these heads' lines - 20 BOMs' worth instead of all 130k.
	if err := services.DbGetNoCache(&response.ItemDetails, map[string]interface{}{"item_bom_id": ids}); err != nil {
		return response, pagination, fiber.StatusInternalServerError, errors.New("failed getting bom details")
	}

	return response, pagination, 0, nil
}

// GetBomsSearch backs the BOM screen's Search: matching heads only, 20 to a page, numbered, read
// down by name then model. The modal hands back the chosen BOM's id and the screen opens the
// page starting at it.
func GetBomsSearch(search string, page int) ([]models.BomViewList, utils.PaginationMeta, int, error) {
	var rows []models.BomViewList

	if page < 1 {
		page = 1
	}

	searchColumns := []string{"item_code", "general_name", "item_model", "short_desc"}

	total, err := services.DbSearchPage(&rows, nil, search, searchColumns, nil, page, bomPageSize,
		"general_name ASC, item_model ASC, id ASC")
	if err != nil {
		return rows, utils.PaginationMeta{}, fiber.StatusInternalServerError, errors.New("failed searching boms")
	}

	totalPages := int((total + int64(bomPageSize) - 1) / int64(bomPageSize))

	return rows, utils.PaginationMeta{
		HasNext:    page < totalPages,
		HasPrev:    page > 1,
		PageSize:   bomPageSize,
		Page:       page,
		TotalPages: totalPages,
		Total:      total,
	}, 0, nil
}

// BomTreeMaxIDs bounds one request; bomTreeChunk keeps every IN list far under SQL Server's
// 2,100-parameter limit however wide a tree fans out.
const (
	BomTreeMaxIDs = 500
	bomTreeChunk  = 500
)

// GetBomTree serves the quotation screens (Sales and Engineering) one BOM at a time: the BOMs
// asked for, each with its component lines, and every sub-BOM under them - a component whose
// item has a BOM of its own - to any depth, in the same bom_head/bom_details shape as GET
// /setup/bom. The screens used to download GET /setup/bom - every BOM head and line in the
// system, 32.6 MB of JSON on test_fresh - on every quotation opened, to expand the one or two
// BOMs a user picks (user request 2026-09-26: "optimize the BOM ... also update the
// engineering side"). Uncached, like GetBomsPaged: a few indexed seeks per level.
func GetBomTree(ids []int) (BomPageResponse, int, error) {
	response := BomPageResponse{ItemBom: []models.BomViewList{}, ItemDetails: []models.BomViewItemList{}}

	seen := map[uint]bool{}
	frontier := []uint{}
	for _, id := range ids {
		if id > 0 && !seen[uint(id)] {
			seen[uint(id)] = true
			frontier = append(frontier, uint(id))
		}
	}

	for len(frontier) > 0 {
		var heads []models.BomViewList
		var lines []models.BomViewItemList
		for _, part := range chunkUint(frontier, bomTreeChunk) {
			var h []models.BomViewList
			if err := services.DbGetNoCache(&h, map[string]interface{}{"id": part}); err != nil {
				return response, fiber.StatusInternalServerError, errors.New("failed getting bom heads")
			}
			heads = append(heads, h...)

			var l []models.BomViewItemList
			if err := services.DbGetNoCache(&l, map[string]interface{}{"item_bom_id": part}); err != nil {
				return response, fiber.StatusInternalServerError, errors.New("failed getting bom details")
			}
			lines = append(lines, l...)
		}
		response.ItemBom = append(response.ItemBom, heads...)
		response.ItemDetails = append(response.ItemDetails, lines...)

		// The next level: BOMs of the components just read.
		itemSeen := map[uint]bool{}
		components := []uint{}
		for _, line := range lines {
			if line.ItemID > 0 && !itemSeen[line.ItemID] {
				itemSeen[line.ItemID] = true
				components = append(components, line.ItemID)
			}
		}

		frontier = frontier[:0]
		for _, part := range chunkUint(components, bomTreeChunk) {
			var subs []models.BomViewList
			if err := services.DbGetNoCache(&subs, map[string]interface{}{"item_id": part}); err != nil {
				return response, fiber.StatusInternalServerError, errors.New("failed getting sub-boms")
			}
			for _, sub := range subs {
				if !seen[sub.ID] {
					seen[sub.ID] = true
					frontier = append(frontier, sub.ID)
				}
			}
		}
	}

	return response, 0, nil
}

func chunkUint(ids []uint, size int) [][]uint {
	var chunks [][]uint
	for start := 0; start < len(ids); start += size {
		end := start + size
		if end > len(ids) {
			end = len(ids)
		}
		chunks = append(chunks, ids[start:end])
	}
	return chunks
}
