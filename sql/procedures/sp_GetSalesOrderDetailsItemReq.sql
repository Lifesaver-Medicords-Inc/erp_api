ALTER PROCEDURE [dbo].[sp_GetSalesOrderDetailsItemReq] @SalesId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT sod.order_details_id AS sales_order_details_id,
    so.order_id AS sales_order_id,
    sod.item_id,
    sod.item_description AS item_desc,
    sod.qty AS required_qty,
    uom.name AS required_uom,
    -- remaining order qty
    CASE
        WHEN (sod.qty - ISNULL(SUM(ird.required_qty), 0)) < 0 THEN 0
        ELSE (sod.qty - ISNULL(SUM(ird.required_qty), 0))
    END AS remaining_qty,
    uom.name AS remaining_uom
FROM tbl_trans_sales_order_details sod
    LEFT JOIN tbl_inv_item_request_details2 ird ON ird.sales_order_details_id = sod.order_details_id
    INNER JOIN tbl_trans_sales_order so ON sod.based_id = so.order_id
    INNER JOIN tbl_setup_item i ON sod.item_id = i.id
    INNER JOIN tbl_setup_item_unit_measurement uom ON i.unit_of_measure_id = uom.id
WHERE sod.based_id = @SalesId
GROUP BY sod.order_details_id,
    sod.item_id,
    sod.item_code,
    sod.item_description,
    sod.qty,
    uom.name,
    so.order_id
HAVING (sod.qty - ISNULL(SUM(ird.required_qty), 0)) > 0;
END TRY BEGIN CATCH THROW;
END CATCH
END;