-- §7.1 row 15 (FOR DELIVERY): recompute the affected SO line's status whenever a
-- Delivery Receipt line item is created. sales_order_details_id is a direct FK. Rows
-- 16-17 (FAILED/RETURN, DELIVERED) fire from tr_so_item_status_from_logistics_route
-- instead, since delivery outcome lives on the logistics route, not the DR itself.
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_dr]
ON [dbo].[tbl_dispatching_delivery_receipt_items]
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @id BIGINT;
    DECLARE cur CURSOR LOCAL FAST_FORWARD FOR
        SELECT DISTINCT sales_order_details_id FROM inserted WHERE sales_order_details_id IS NOT NULL;

    OPEN cur;
    FETCH NEXT FROM cur INTO @id;
    WHILE @@FETCH_STATUS = 0
    BEGIN
        EXEC sp_RecomputeSoItemStatus @order_details_id = @id;
        FETCH NEXT FROM cur INTO @id;
    END
    CLOSE cur;
    DEALLOCATE cur;
END;
