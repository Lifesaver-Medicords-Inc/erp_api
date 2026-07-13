CREATE OR ALTER VIEW [dbo].[vw_get_sales_order_with_approved_ir] AS
SELECT
	a.order_id as sales_order_id,
    a.customer_id, 
    c.branch_name AS customer_name,
	c.customer_code,
    e.location as address,
	a.ship_type_id,
	d.tin as tin_no,
	a.ship_to_id,
    a.doc as sales_order_no,
	b.id as item_release_id,
    b.doc_no as item_release_no,
    b.reference_doc_no,
    f.location as deliver_to,
	b.required_date as delivery_date,
	a.sales_executive
FROM tbl_trans_sales_order a
LEFT JOIN tbl_inv_item_release b ON a.doc = b.reference_doc_no
LEFT JOIN tbl_bpi_general c ON a.customer_id = c.id 
LEFT JOIN tbl_bpi d ON c.based_id = d.id
OUTER APPLY (
    SELECT TOP 1 location
    FROM tbl_bpi_address
    WHERE branch_id = a.customer_id
    ORDER BY id ASC
) e
OUTER APPLY (
    SELECT
	location
    FROM tbl_bpi_address
    WHERE id = a.ship_to_id
) f
WHERE b.approved_by IS NOT NULL AND b.approved_by <> ''