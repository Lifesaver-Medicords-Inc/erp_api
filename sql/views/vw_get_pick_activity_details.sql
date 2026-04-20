ALTER VIEW [dbo].[vw_get_pick_activity_details] AS
SELECT pad.*,
    -- Total actual qty per Sales Order detail
    ISNULL(
        SUM(pad.actual_qty) OVER (PARTITION BY pad.sales_order_details_id),
        0
    ) AS total_actual_qty,
    -- Left qty = original qty - total actual qty
    sod.qty - ISNULL(
        SUM(pad.actual_qty) OVER (PARTITION BY pad.sales_order_details_id),
        0
    ) AS left_qty,
    pad.actual_uom AS left_uom
FROM tbl_inv_pick_activity_details2 pad
    INNER JOIN tbl_trans_sales_order_details sod ON pad.sales_order_details_id = sod.order_details_id;