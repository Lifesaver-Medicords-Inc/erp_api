CREATE PROCEDURE [dbo].[sp_GetPurchaseOrderDetailsInvoice]
    @PurchaseId INT
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        SELECT
		    pod.id,
			pod.based_id,
			pod.item_code,
			pod.item_description,
			pod.unit_of_measure AS order_uom,
			pod.order_qty,
			pod.unit_of_measure AS req_uom,
			pod.req_qty AS item_qty,
			pod.discounted_price AS unit_price,
			pod.total_price AS total_cost
        FROM dbo.tbl_accounting_invoice_receipt_details AS ird
		LEFT JOIN tbl_purchasing_purchase_order_details AS pod
			ON ird.purchase_order_details_id = pod.id
		LEFT JOIN tbl_purchasing_purchase_order AS po
			ON pod.based_id = po.id
        WHERE pod.based_id = @PurchaseId;
    END TRY
    BEGIN CATCH
        THROW;
    END CATCH
END;
GO


