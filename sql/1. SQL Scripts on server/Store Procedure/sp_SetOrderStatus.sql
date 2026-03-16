CREATE procedure [dbo].[sp_SetOrderStatus] 
	@OrderDetailId int, @OrderType nvarchar(5)
AS
-- Update sales order details if OrderType is 'SO'
IF @OrderType = 'SO'
	BEGIN
		UPDATE tbl_trans_sales_order_details
		SET status = CASE 
				WHEN allocated_qty >= qty THEN 'WAITING FOR DELIVERY'
				ELSE 'CANVASS'
			END
		WHERE order_details_id = @OrderDetailId
	END
-- Update sales order details if OrderType is 'PR'
ELSE IF @OrderType = 'PR'
	BEGIN
		UPDATE tbl_purchasing_purchase_requisition_orders
		SET status = CASE
		WHEN allocated_qty >= qty THEN 'WAITING FOR DELIVERY'
				ELSE 'CANVASS'
			END
		WHERE pr_order_id = @OrderDetailId
	END
GO


