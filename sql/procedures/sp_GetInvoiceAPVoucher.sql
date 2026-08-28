ALTER PROCEDURE [dbo].[sp_GetInvoiceAPVoucher] @SupplierId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
-- open_amount = net_amount minus everything already applied against this
-- receipt via any AP Voucher or Debit Memo (see services.ComputeReceiptOpenAmount,
-- the Go-side equivalent used for save-time validation - this proc is the
-- read/picker side of the same calculation, kept in sync deliberately since
-- both AP Voucher and Debit Memo share this one picker). Replaces the old
-- ISNULL(ap_voucher, 0) = 0 boolean filter, which only ever supported an
-- all-or-nothing "already fully vouchered" state - a receipt partially
-- applied by one APV or DM still has ap_voucher = 1 today but genuinely has
-- room left, and used to disappear from this list entirely.
WITH ir_open AS (
    SELECT a.id AS invoice_receipt_id,
        a.doc_no AS receipt_no,
        a.doc_date AS ir_doc_date,
        a.invoice_due AS ir_due_date,
        a.twas_amount AS twas_amount,
        a.net_amount AS line_amount,
        'INVOICE RECEIPT' AS receipt_type,
        a.net_amount
            - ISNULL((
                SELECT SUM(apd.amount_applied) FROM tbl_accounting_ap_voucher_details apd
                WHERE apd.invoice_receipt_id = a.id AND apd.receipt_type = 'INVOICE RECEIPT'
            ), 0)
            - ISNULL((
                SELECT SUM(dmd.amount_applied) FROM tbl_trans_debit_memo_details dmd
                WHERE dmd.target_doc_id = a.id AND dmd.target_doc_type = 'Invoice Receipt' AND dmd.apply = 1
            ), 0) AS open_amount
    FROM tbl_accounting_invoice_receipt a
    WHERE a.supplier_id = @SupplierId

    UNION ALL

    SELECT b.id AS invoice_receipt_id,
        b.doc_no AS receipt_no,
        b.doc_date AS ir_doc_date,
        b.invoice_due AS ir_due_date,
        b.twas_amount AS twas_amount,
        b.net_amount AS line_amount,
        'BULK INVOICE RECEIPT' AS receipt_type,
        b.net_amount
            - ISNULL((
                SELECT SUM(apd.amount_applied) FROM tbl_accounting_ap_voucher_details apd
                WHERE apd.invoice_receipt_id = b.id AND apd.receipt_type = 'BULK INVOICE RECEIPT'
            ), 0)
            - ISNULL((
                SELECT SUM(dmd.amount_applied) FROM tbl_trans_debit_memo_details dmd
                WHERE dmd.target_doc_id = b.id AND dmd.target_doc_type = 'Bulk Invoice Receipt' AND dmd.apply = 1
            ), 0) AS open_amount
    FROM tbl_accounting_bulk_invoice_receipt b
    WHERE b.supplier_id = @SupplierId
)
SELECT invoice_receipt_id, receipt_no, ir_doc_date, ir_due_date, twas_amount,
    line_amount, receipt_type, open_amount
FROM ir_open
WHERE open_amount > 0.005;
END TRY BEGIN CATCH THROW;
END CATCH
END;
