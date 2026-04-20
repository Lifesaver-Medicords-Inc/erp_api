ALTER VIEW [dbo].[vw_get_sales_order_item_req_doc] AS
SELECT DISTINCT so.order_id AS sales_order_id,
    so.doc AS so_doc_no
FROM tbl_trans_sales_order so
    INNER JOIN tbl_trans_sales_order_details sod ON so.order_id = sod.based_id
    LEFT JOIN (
        SELECT ird.sales_order_details_id,
            SUM(ird.required_qty) AS total_required_qty
        FROM tbl_inv_item_request2 ir
            INNER JOIN tbl_inv_item_request_details2 ird ON ir.id = ird.item_request_id
        GROUP BY ird.sales_order_details_id
    ) agg_ird ON sod.order_details_id = agg_ird.sales_order_details_id
WHERE (
        sod.qty - ISNULL(agg_ird.total_required_qty, 0)
    ) > 0