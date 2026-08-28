package services

import (
	"gorm.io/gorm"
)

// ComputeReceiptOpenAmount is the shared "how much is actually left on this
// Invoice Receipt / Bulk Invoice Receipt" calculation, per spec §5.16/§5.19 -
// both AP Voucher and Debit Memo apply against the SAME open amount, so this
// lives here rather than duplicated in either service.
//
// Deliberately computed live (net_amount minus every AP Voucher detail and
// every applied Debit Memo detail ever recorded against this receipt) rather
// than stored as a column that both callers would need to remember to
// decrement - the same pattern already used one level down, where
// vw_get_payment_voucher_details computes a Payment Voucher's open amount
// against an AP Voucher the same way (SUM(...) OVER, not a stored balance).
// A stored/decremented column would need its own restore-on-delete and
// restore-on-edit handling (see ERP_API's Item Release fix for how much that
// costs); computing it fresh every time needs none of that - deleting or
// editing an AP Voucher or Debit Memo detail row just changes what the next
// SUM sees.
//
// receiptType must be "INVOICE RECEIPT" or "BULK INVOICE RECEIPT" (the same
// two values sp_GetInvoiceAPVoucher's own receipt_type column uses).
func ComputeReceiptOpenAmount(tx *gorm.DB, receiptType string, receiptId uint) (float64, error) {
	var netAmount float64
	var table string
	switch receiptType {
	case "INVOICE RECEIPT":
		table = "tbl_accounting_invoice_receipt"
	case "BULK INVOICE RECEIPT":
		table = "tbl_accounting_bulk_invoice_receipt"
	default:
		return 0, gorm.ErrInvalidData
	}

	if err := tx.Raw("SELECT net_amount FROM "+table+" WHERE id = ?", receiptId).Scan(&netAmount).Error; err != nil {
		return 0, err
	}

	var appliedViaApv float64
	if err := tx.Raw(`
		SELECT ISNULL(SUM(amount_applied), 0)
		FROM tbl_accounting_ap_voucher_details
		WHERE invoice_receipt_id = ? AND receipt_type = ?
	`, receiptId, receiptType).Scan(&appliedViaApv).Error; err != nil {
		return 0, err
	}

	// DM's own target_doc_type spelling ("Invoice Receipt"/"Bulk Invoice
	// Receipt") differs in case/format from receipt_type's all-caps form
	// used everywhere else - map it explicitly rather than assuming they
	// match, since they never have.
	dmTargetType := "Invoice Receipt"
	if receiptType == "BULK INVOICE RECEIPT" {
		dmTargetType = "Bulk Invoice Receipt"
	}

	var appliedViaDm float64
	if err := tx.Raw(`
		SELECT ISNULL(SUM(amount_applied), 0)
		FROM tbl_trans_debit_memo_details
		WHERE target_doc_id = ? AND target_doc_type = ? AND apply = 1
	`, receiptId, dmTargetType).Scan(&appliedViaDm).Error; err != nil {
		return 0, err
	}

	return netAmount - appliedViaApv - appliedViaDm, nil
}

// ComputeCreditMemoOpenAmount is the same idea for a supplier Credit Memo
// being used as a Debit Memo apply target (§5.19's third target type) -
// trans_amount minus every Debit Memo detail already applied against it.
func ComputeCreditMemoOpenAmount(tx *gorm.DB, creditMemoId uint) (float64, error) {
	var transAmount float64
	if err := tx.Raw(`
		SELECT trans_amount FROM tbl_trans_credit_memo WHERE id = ?
	`, creditMemoId).Scan(&transAmount).Error; err != nil {
		return 0, err
	}

	var appliedViaDm float64
	if err := tx.Raw(`
		SELECT ISNULL(SUM(amount_applied), 0)
		FROM tbl_trans_debit_memo_details
		WHERE target_doc_id = ? AND target_doc_type = 'Credit Memo' AND apply = 1
	`, creditMemoId).Scan(&appliedViaDm).Error; err != nil {
		return 0, err
	}

	return transAmount - appliedViaDm, nil
}
