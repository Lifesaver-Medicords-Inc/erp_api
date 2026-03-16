CREATE VIEW [dbo].[vw_get_supplier_trade] AS 
SELECT
   a.id AS supplier_id,
   a.branch_name AS supplier,
   a.supplier_code,
   UPPER(a.transaction_type) AS invoice_type,
   c.name AS payment_term,
   'INVOICE TYPE' AS [type]
FROM tbl_bpi_general a
LEFT JOIN tbl_bpi_finance b
	ON a.id = b.finance_id
INNER JOIN tbl_setup_payment_terms c
	ON b.finance_payment_terms_id = c.id
GO


