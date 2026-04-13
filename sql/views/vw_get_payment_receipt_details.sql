ALTER VIEW [dbo].[vw_get_payment_receipt_details] AS
SELECT prd.id,
    prd.payment_receipt_id,
    prd.sales_invoice_id,
    prd.doc_no,
    prd.due_date,
    prd.doc_date,
    prd.amount_applied,
    prd.twas_applied,
    -- Total applied per Sales Invoice detail
    ISNULL(
        SUM(prd.amount_applied) OVER (PARTITION BY prd.sales_invoice_id),
        0
    ) AS total_amount_applied,
    -- Open amount = original total_cost - total applied
    si.total_amount_due - ISNULL(
        SUM(prd.amount_applied) OVER (PARTITION BY prd.sales_invoice_id),
        0
    ) AS open_amount
FROM tbl_accounting_payment_receipt_details prd
    INNER JOIN tbl_accounting_sales_invoice si ON prd.sales_invoice_id = si.id;