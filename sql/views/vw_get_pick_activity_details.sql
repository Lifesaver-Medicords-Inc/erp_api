ALTER VIEW [dbo].[vw_get_pick_activity_details] AS
SELECT pad.*,
    -- Total pick qty per Sales Order detail
    ISNULL(
        SUM(pad.pick_qty) OVER (PARTITION BY pad.sales_order_details_id),
        0
    ) AS total_pick_qty,
    -- Left qty = original qty - total pick qty
    sod.qty - ISNULL(
        SUM(pad.pick_qty) OVER (PARTITION BY pad.sales_order_details_id),
        0
    ) AS left_qty,
    pad.pick_uom AS left_uom
FROM tbl_inv_pick_activity_details2 pad
    LEFT JOIN tbl_trans_sales_order_details sod ON pad.sales_order_details_id = sod.order_details_id;