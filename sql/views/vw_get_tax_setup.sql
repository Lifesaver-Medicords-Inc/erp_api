CREATE
OR ALTER VIEW [dbo].[vw_get_tax_setup] AS
SELECT a.id AS view_id,
    a.code,
    a.tax_desc,
    a.coa_purchase_id AS coa_purch_id,
    b.name AS output_tax_account,
    a.coa_sales_id AS coa_sales_id,
    c.name AS input_tax_account,
    d.tax_rate,
    CASE
        WHEN d.valid_to IS NULL
        OR LTRIM(RTRIM(d.valid_to)) = '' THEN CONCAT(d.valid_from, ' - INDEFINITE')
        ELSE CONCAT(d.valid_from, ' - ', d.valid_to)
    END AS effective_period,
    COALESCE(e.type, f.type) AS account_type
FROM tbl_setup_tax a
    LEFT JOIN tbl_setup_chart_of_accounts b ON a.coa_sales_id = b.id
    LEFT JOIN tbl_setup_chart_of_accounts c ON a.coa_purchase_id = c.id
    LEFT JOIN vw_get_tax_setup_details d ON a.id = d.tax_code_id
    LEFT JOIN tbl_setup_chart_class e ON b.class_id = e.id
    LEFT JOIN tbl_setup_chart_class f ON c.class_id = f.id
WHERE d.valid_status = 'ACTIVE'