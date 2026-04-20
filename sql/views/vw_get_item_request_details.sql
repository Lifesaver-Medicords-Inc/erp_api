ALTER VIEW [dbo].[vw_get_item_request_details] AS
SELECT ird.*,
    -- Total required qty per Sales Order detail
    ISNULL(
        SUM(ird.required_qty) OVER (PARTITION BY ird.sales_order_details_id),
        0
    ) AS total_required_qty,
    -- Remaining qty = original qty - total required qty
    sod.qty - ISNULL(
        SUM(ird.required_qty) OVER (PARTITION BY ird.sales_order_details_id),
        0
    ) AS remaining_qty,
    ird.required_uom AS remaining_uom
FROM tbl_inv_item_request_details2 ird
    INNER JOIN tbl_trans_sales_order_details sod ON ird.sales_order_details_id = sod.order_details_id;