CREATE PROCEDURE [dbo].[sp_GetPurchaseOrders]
    @PoId INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        SELECT 
            po.id AS purchase_order_id,
            po.supplier_id,
            po.supplier_name,
            po.supplier_code,
            po.doc_no,
            pod.id AS pod_id,
            pod.item_id,
			pod.order_qty,
            ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS total_received_qty,
            pod.order_qty - ISNULL(SUM(CAST(wh.received_qty AS INT)), 0) AS ordered_qty,
			pod.unit_of_measure AS ordered_uom,
			pod.item_code,
			ias.long_description AS item_description
        FROM tbl_inv_warehouse_receiving_history wh
		LEFT JOIN tbl_purchasing_purchase_order po
            ON wh.purchase_order_id = wh.purchase_order_id
        INNER JOIN tbl_purchasing_purchase_order_details pod
            ON wh.purchase_order_details_id = pod.id
		INNER JOIN tbl_setup_item i
			ON pod.item_id = i.id
		INNER JOIN tbl_setup_item_additional_specs ias
			ON i.id = ias.based_id
        WHERE po.id = @PoId
        GROUP BY 
            po.id,
            po.supplier_id,
            po.supplier_name,
            po.supplier_code,
            po.doc_no,
            pod.id,
            pod.item_id,
            pod.order_qty,
			pod.unit_of_measure,
			pod.item_code,
			ias.long_description
    END TRY
    BEGIN CATCH
        THROW;
    END CATCH
END;
GO


