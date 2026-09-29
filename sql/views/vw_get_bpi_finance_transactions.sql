-- BPI Finance tab, TRANSACTIONS (spec 4.1.7): EVERY sales order of the partner, one row each
-- (user instruction 2026-09-29 - until then only the finished ones: invoiced, CLOSED or
-- CANCELLED), with what was raised against it: the POs (stacked in the row, one per line),
-- the sales invoice with its doc date, and the Payment Receipt (PAYR#, the in-system
-- settlement document - not the Official Receipt, which Billing tracks instead) with its date.
-- Dates are the documents' own. Added 2026-09-26: the tab had columns and nothing filled them.
-- Each row also carries what the SO is worth, what has been paid against it and the difference.
-- The partner's ACCOUNT BALANCE is the sum of that BALANCE column (user instruction 2026-09-29:
-- "all sales order - payment in sales invoice (account receivables)"):
--   * TOTAL AMOUNT DUE is the SO's own. A CANCELLED SO is owed only what was invoiced on it -
--     the cancellation charges, if any were approved (spec 5.4, 5.20) - not its old total.
--   * PAID is every Payment Receipt line applied to the SO's invoices, the withholding tax the
--     customer kept included - the same sum Billing's A/R uses (billing_service.go).
-- An invoice belongs to the SO its SO REF names, compared the way Billing compares it - the
-- typed reference may read SO#0012, SO 0012 or 0012.
-- SI and PAYR numbers are stored as integers and shown as SI#0001 / PAYR#0001 (spec 2.5:
-- four digits, wider once past 9999).
-- The stacked columns are built with FOR XML PATH ... .value(), which SQL Server allows only
-- under QUOTED_IDENTIFIER ON - so the view is created with it on, whatever runs this file.
SET QUOTED_IDENTIFIER ON;
SET ANSI_NULLS ON;
GO
IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_finance_transactions' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_finance_transactions] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_finance_transactions] AS
SELECT so.order_id,
    'SO#' + so.doc AS sales_order,
    CONVERT(varchar(10), TRY_CONVERT(datetime, so.date), 101) AS sales_order_date,
    po.purchase_orders,
    po.purchase_order_dates,
    si.sales_invoices,
    si.sales_invoice_dates,
    pr.payment_receipts,
    pr.payment_receipt_dates,
    CAST(amt.total_amount_due AS decimal(18, 2)) AS total_amount_due,
    CAST(ISNULL(paid.paid, 0) AS decimal(18, 2)) AS paid,
    CAST(amt.total_amount_due - ISNULL(paid.paid, 0) AS decimal(18, 2)) AS balance,
    so.status,
    so.customer_id,
    d.id AS finance_transaction_branch_id
FROM tbl_trans_sales_order so
    OUTER APPLY (
        SELECT STUFF((SELECT CHAR(13) + CHAR(10) + 'PO#' + p.doc_no
                        FROM tbl_purchasing_purchase_order p
                       WHERE p.ref_doc_no = 'SO#' + so.doc
                       ORDER BY p.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS purchase_orders,
               STUFF((SELECT CHAR(13) + CHAR(10) + ISNULL(CONVERT(varchar(10), TRY_CONVERT(datetime, p.date), 101), '')
                        FROM tbl_purchasing_purchase_order p
                       WHERE p.ref_doc_no = 'SO#' + so.doc
                       ORDER BY p.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS purchase_order_dates
    ) po
    OUTER APPLY (
        SELECT STUFF((SELECT CHAR(13) + CHAR(10) + 'SI#' + CASE WHEN i.doc_no > 9999 THEN CAST(i.doc_no AS varchar(20)) ELSE RIGHT('0000' + CAST(i.doc_no AS varchar(20)), 4) END
                        FROM tbl_accounting_sales_invoice i
                       WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc
                       ORDER BY i.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS sales_invoices,
               STUFF((SELECT CHAR(13) + CHAR(10) + ISNULL(i.doc_date, '')
                        FROM tbl_accounting_sales_invoice i
                       WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc
                       ORDER BY i.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS sales_invoice_dates
    ) si
    -- Each receipt once, with ITS date on the matching line: a receipt settling two invoices
    -- of the same SO is one receipt, and two receipts of the same day are still two lines.
    OUTER APPLY (
        SELECT STUFF((SELECT CHAR(13) + CHAR(10) + 'PAYR#' + CASE WHEN x.doc_no > 9999 THEN CAST(x.doc_no AS varchar(20))
                                                      ELSE RIGHT('0000' + CAST(x.doc_no AS varchar(20)), 4) END
                        FROM (SELECT DISTINCT r.id, r.doc_no
                                FROM tbl_accounting_payment_receipt r
                                JOIN tbl_accounting_payment_receipt_details rd ON rd.payment_receipt_id = r.id
                                JOIN tbl_accounting_sales_invoice i ON i.id = rd.sales_invoice_id
                               WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc) x
                       ORDER BY x.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS payment_receipts,
               STUFF((SELECT CHAR(13) + CHAR(10) + ISNULL(x.doc_date, '')
                        FROM (SELECT DISTINCT r.id, r.doc_date
                                FROM tbl_accounting_payment_receipt r
                                JOIN tbl_accounting_payment_receipt_details rd ON rd.payment_receipt_id = r.id
                                JOIN tbl_accounting_sales_invoice i ON i.id = rd.sales_invoice_id
                               WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc) x
                       ORDER BY x.id
                         FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '') AS payment_receipt_dates
    ) pr
    OUTER APPLY (
        SELECT SUM(ISNULL(i.total_amount_due, 0)) AS invoiced
          FROM tbl_accounting_sales_invoice i
         WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc
    ) inv
    OUTER APPLY (
        SELECT CASE WHEN ISNULL(so.status, '') = 'CANCELLED' THEN ISNULL(inv.invoiced, 0)
                    ELSE ISNULL(so.total_amount_due, 0) END AS total_amount_due
    ) amt
    OUTER APPLY (
        SELECT SUM(ISNULL(rd.amount_applied, 0) + ISNULL(rd.twas_applied, 0)) AS paid
          FROM tbl_accounting_payment_receipt_details rd
          JOIN tbl_accounting_sales_invoice i ON i.id = rd.sales_invoice_id
         WHERE REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(i.reference_doc_so, '')))), 'SO#', ''), 'SO', ''), ' ', '') = so.doc
    ) paid
    LEFT JOIN (
        SELECT based_id, MIN(id) AS id
          FROM tbl_bpi_general
         GROUP BY based_id
    ) d ON so.customer_id = d.based_id;
