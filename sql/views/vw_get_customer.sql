ALTER VIEW [dbo].[vw_get_customer] AS
SELECT a.based_id AS customer_id,
    a.branch_name AS customer,
    a.customer_code,
    c.name AS payment_term,
    b.finance_tax_code AS tax_code,
    b.finance_tax AS tax,
    e.location AS customer_address,
    bpi.tin,
    CASE
        WHEN ISNULL(SUM(d.bpi_overpayment_amount), 0) < 0.01 THEN 0
        ELSE ISNULL(SUM(d.bpi_overpayment_amount), 0)
    END AS overpayment_amount
FROM tbl_bpi_general a
    LEFT JOIN tbl_bpi_finance b ON a.based_id = b.finance_based_id
    LEFT JOIN tbl_setup_payment_terms c ON b.finance_payment_terms_id = c.id
    LEFT JOIN tbl_accounting_bpi_overpayment d ON a.id = d.bpi_id
    LEFT JOIN tbl_bpi_address e ON a.id = e.branch_id
    INNER JOIN tbl_bpi bpi ON a.based_id = bpi.id
WHERE a.customer_code IS NOT NULL
    AND LTRIM(RTRIM(a.customer_code)) <> ''
GROUP BY a.based_id,
    a.branch_name,
    a.customer_code,
    c.name,
    b.finance_tax_code,
    b.finance_tax,
    e.location,
    bpi.tin