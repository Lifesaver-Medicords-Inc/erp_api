ALTER VIEW [dbo].[vw_get_sales_order_pick_act_doc] AS
SELECT DISTINCT so.order_id AS sales_order_id,
    so.doc AS so_doc_no
FROM tbl_trans_sales_order so
    INNER JOIN tbl_trans_sales_order_details sod ON so.order_id = sod.based_id
    LEFT JOIN (
        SELECT drd.sales_order_details_id,
            SUM(drd.qty) AS total_delivered_qty
        FROM tbl_dispatching_delivery_receipt dr
            INNER JOIN tbl_dispatching_delivery_receipt_items drd ON dr.id = drd.sales_order_details_id
        GROUP BY drd.sales_order_details_id
    ) agg_drd ON sod.order_details_id = agg_drd.sales_order_details_id
WHERE (
        sod.qty - ISNULL(agg_drd.total_delivered_qty, 0)
    ) > 0