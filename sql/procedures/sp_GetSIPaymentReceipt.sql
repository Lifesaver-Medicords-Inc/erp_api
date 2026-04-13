CREATE
OR ALTER PROCEDURE [dbo].[sp_GetSIPaymentReceipt] @CustomerId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT si.id AS sales_invoice_id,
    si.doc_no,
    si.doc_date,
    si.posting_date AS due_date,
    si.less_vat AS twas_applied,
    si.total_amount_due - CASE
        WHEN ABS(ISNULL(SUM(pvd.amount_applied), 0)) < 0.01 THEN 0
        ELSE ISNULL(SUM(pvd.amount_applied), 0)
    END AS open_amount
FROM dbo.tbl_accounting_sales_invoice AS si
    INNER JOIN dbo.tbl_accounting_payment_receipt_details pvd ON si.id = pvd.sales_invoice_id
WHERE si.customer_id = @CustomerId
GROUP BY si.id,
    si.doc_no,
    si.doc_date,
    si.total_amount_due,
    si.posting_date,
    si.less_vat
HAVING si.total_amount_due - CASE
        WHEN ABS(ISNULL(SUM(pvd.amount_applied), 0)) < 0.01 THEN 0
        ELSE ISNULL(SUM(pvd.amount_applied), 0)
    END > 0
END TRY BEGIN CATCH THROW;
END CATCH
END;