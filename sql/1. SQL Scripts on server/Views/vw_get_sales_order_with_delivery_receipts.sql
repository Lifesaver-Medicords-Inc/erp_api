CREATE VIEW [dbo].[vw_get_sales_order_with_delivery_receipts] AS 
SELECT
	a.order_id,
	a.customer_id,	
	c.branch_name AS customer_name,
	c.customer_code,
	d.finance_tax_code AS tax_code,
	d.finance_tax as tax_id,
	a.project_name AS address,
	a.payment_terms_id,
	a.ref_po,
	a.document_no AS doc_so_no,
	a.date AS doc_date,
	a.total_amount_due AS net_amount,
	a.sales_executive
FROM tbl_trans_sales_order a 
LEFT JOIN tbl_bpi_general c ON a.customer_id = c.id
LEFT JOIN tbl_bpi_finance d ON a.customer_id = d.finance_branch_id


GO
