CREATE VIEW  [dbo].[vw_qqnotes] AS 

  SELECT
	a.id as qq_id,
	a.based_id,
	a.components,
	a.model,
	a.item_id,
	a.bom_id,
	a.is_child,
	a.qty,
	a.unit_of_measure,
	b.remarks,
	b.notes,
	b.id as qq_note_id,
	d.name as customer_name

 FROM tbl_trans_sales_quotation_quick  a
 LEFT JOIN tbl_setup_item_boq_details b ON a.id = b.qq_id
 LEFT JOIN tbl_trans_sales_quotation c ON a.based_id = c.id
 LEFT JOIN tbl_bpi d ON c.customer_id =  d.id


GO
