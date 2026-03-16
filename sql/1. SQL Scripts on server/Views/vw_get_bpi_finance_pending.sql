CREATE VIEW [dbo].[vw_get_bpi_finance_pending] AS
SELECT 
	a.date,
    b.document_no as qoute_ref,
	a.gross_sales as total_price,
	b.stage,
	b.status,
	a.customer_id,
	d.id as finance_pending_branch_id
FROM tbl_trans_sales_quotation a
LEFT JOIN tbl_trans_sales_opportunity b ON a.document_no = b.document_no

LEFT JOIN (
     SELECT 
       based_id, 
       MIN(id) as id
   FROM tbl_bpi_general
   GROUP BY based_id
) d ON a.customer_id = d.based_id;


GO
