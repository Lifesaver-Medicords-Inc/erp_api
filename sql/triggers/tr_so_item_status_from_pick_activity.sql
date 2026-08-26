-- §7.1 rows 11-12: recompute the affected SO line's status whenever a pick activity
-- line is created or its actual_qty changes. sales_order_details_id is a direct FK.
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_pick_activity]
ON [dbo].[tbl_inv_pick_activity_details2]
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
