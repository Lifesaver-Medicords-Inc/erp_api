ALTER PROCEDURE [dbo].[sp_GetInvoiceAPVoucher] @SupplierId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY -- Invoice Receipt
SELECT a.id AS invoice_receipt_id,
    a.doc_no AS receipt_no,
    a.doc_date AS ir_doc_date,
    a.invoice_due AS ir_due_date,
    a.twas_amount AS twas_amount,
    a.net_amount AS line_amount,
    'INVOICE RECEIPT' AS receipt_type
FROM tbl_accounting_invoice_receipt a
WHERE a.supplier_id = @SupplierId
    AND ISNULL(a.ap_voucher, 0) = 0
UNION ALL
-- Bulk Invoice Receipt
SELECT b.id AS invoice_receipt_id,
    b.doc_no AS receipt_no,
    b.invoice_due AS ir_due_date,
    b.doc_date AS ir_doc_date,
    b.twas_amount AS twas_amount,
    b.net_amount AS line_amount,
    'BULK INVOICE RECEIPT' AS receipt_type
FROM tbl_accounting_bulk_invoice_receipt b
WHERE b.supplier_id = @SupplierId
    AND ISNULL(b.ap_voucher, 0) = 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;