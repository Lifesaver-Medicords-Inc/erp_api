ALTER VIEW [dbo].[vw_get_supplier_trade] AS
SELECT a.id AS supplier_id,
    a.branch_name AS supplier,
    a.supplier_code,
    UPPER(a.transaction_type) AS invoice_type,
    c.name AS payment_term,
    'INVOICE TYPE' AS [type],
    e.location AS supplier_address,
    CASE
        WHEN ISNULL(SUM(d.company_overpayment_amount), 0) < 0.01 THEN 0
        ELSE ISNULL(SUM(d.company_overpayment_amount), 0)
    END AS overpayment_amount
FROM tbl_bpi_general a
    LEFT JOIN tbl_bpi_finance b ON a.id = b.finance_id
    LEFT JOIN tbl_setup_payment_terms c ON b.finance_payment_terms_id = c.id
    LEFT JOIN tbl_accounting_bpi_overpayment d ON a.id = d.bpi_id
    LEFT JOIN tbl_bpi_address e ON a.id = e.branch_id
GROUP BY a.id,
    a.branch_name,
    a.supplier_code,
    a.transaction_type,
    c.name,
    e.location;