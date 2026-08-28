package services

import (
	"strconv"
	"strings"

	"gorm.io/gorm"
)

// RecomputeSoItemStatus calls the central §7.1 status-derivation procedure
// (sql/procedures/sp_RecomputeSoItemStatus.sql) for one SO line item. This
// replaces the SQL-trigger approach from ERP_API commit ed10215, reverted in
// 5960c3d: SQL Server unconditionally refuses any INSERT/UPDATE/DELETE with a
// bare OUTPUT clause (not OUTPUT INTO) if the target table has ANY enabled
// trigger, and this codebase's DbInsert/tx.Create() relies on exactly that
// OUTPUT clause almost everywhere to read back a new row's id. Ordinary
// function calls have no such interaction - this is the corrected
// invocation path, called explicitly by each of the seven services whose
// writes affect an SO line's status (Job Order, Pick Activity, Item
// Release, Delivery Receipt, Purchase Order, Receiving Report, Logistics
// Route), right before their own tx.Commit().
//
// The stored procedure itself is unchanged from the trigger-based pass -
// only how it gets invoked changed.
func RecomputeSoItemStatus(tx *gorm.DB, orderDetailsId uint) error {
	if orderDetailsId == 0 {
		return nil
	}

	// §5.25: a repair/replacement pick-up may legitimately cite an already-CLOSED
	// SO - both Item Release's own SO picker and the Logistics Calendar's
	// REFERENCE DOC picker list every SO regardless of status, by design, and
	// neither touches the order's own CLOSED status or anything financial about
	// it. But letting a later document still rewrite that SO's per-line dispatch
	// status (§7.1) is unwanted noise the spec doesn't ask for on a closed order
	// - skip the recompute once the parent SO is CLOSED, so those labels stay
	// frozen at whatever they were the moment the order closed.
	var soStatus string
	if err := tx.Raw(`
		SELECT so.status
		FROM tbl_trans_sales_order_details sod
		INNER JOIN tbl_trans_sales_order so ON so.order_id = sod.based_id
		WHERE sod.order_details_id = ?
	`, orderDetailsId).Scan(&soStatus).Error; err != nil {
		return err
	}
	if strings.EqualFold(strings.TrimSpace(soStatus), "CLOSED") {
		return nil
	}

	return tx.Exec("EXEC sp_RecomputeSoItemStatus @order_details_id = ?", orderDetailsId).Error
}

// RecomputeSoItemStatusForCsv resolves a comma-joined string of
// order_details_id values (tbl_purchasing_purchase_order_details.
// order_detail_ids - one PO line can consolidate the same item across
// several SOs, CLAUDE.md invariant #7) against the real SO detail rows and
// recomputes each one. Same LIKE-based match the stored procedure itself
// uses internally for the reverse direction (this DB's compatibility level,
// 110, has no STRING_SPLIT).
func RecomputeSoItemStatusForCsv(tx *gorm.DB, orderDetailIdsCsv string) error {
	if strings.TrimSpace(orderDetailIdsCsv) == "" {
		return nil
	}

	var ids []uint
	if err := tx.Raw(`
		SELECT sod.order_details_id
		FROM tbl_trans_sales_order_details sod
		WHERE ',' + ? + ',' LIKE '%,' + CAST(sod.order_details_id AS NVARCHAR(20)) + ',%'
	`, orderDetailIdsCsv).Scan(&ids).Error; err != nil {
		return err
	}

	for _, id := range ids {
		if err := RecomputeSoItemStatus(tx, id); err != nil {
			return err
		}
	}
	return nil
}

// RecomputeSoItemStatusForPurchaseOrderDetails looks up one PO detail row's
// own order_detail_ids and recomputes every SO line it consolidates. Used
// by both the PO-creation path (adding a PO line) and the RR-creation path
// (receiving against one).
func RecomputeSoItemStatusForPurchaseOrderDetails(tx *gorm.DB, purchaseOrderDetailsId uint) error {
	if purchaseOrderDetailsId == 0 {
		return nil
	}

	var orderDetailIds string
	if err := tx.Raw(`
		SELECT ISNULL(order_detail_ids, '') FROM tbl_purchasing_purchase_order_details WHERE id = ?
	`, purchaseOrderDetailsId).Scan(&orderDetailIds).Error; err != nil {
		return err
	}

	return RecomputeSoItemStatusForCsv(tx, orderDetailIds)
}

// RecomputeSoItemStatusForDeliveryReceiptDoc resolves a Logistics Route
// leg's delivery_receipt_doc (a doc-number STRING, not a numeric FK - see
// tbl_dispatching_logistics_route) back to the real Delivery Receipt and
// recomputes every SO line on it. doc_no is uniquely indexed, so this is
// reliable despite not being a clean id join.
func RecomputeSoItemStatusForDeliveryReceiptDoc(tx *gorm.DB, deliveryReceiptDoc string) error {
	docNo, err := strconv.Atoi(strings.TrimSpace(deliveryReceiptDoc))
	if err != nil || docNo == 0 {
		return nil
	}

	var ids []uint
	if err := tx.Raw(`
		SELECT DISTINCT dri.sales_order_details_id
		FROM tbl_dispatching_delivery_receipt dr
		INNER JOIN tbl_dispatching_delivery_receipt_items dri ON dri.delivery_receipt_id = dr.id
		WHERE dr.doc_no = ? AND dri.sales_order_details_id IS NOT NULL
	`, docNo).Scan(&ids).Error; err != nil {
		return err
	}

	for _, id := range ids {
		if err := RecomputeSoItemStatus(tx, id); err != nil {
			return err
		}
	}
	return nil
}
