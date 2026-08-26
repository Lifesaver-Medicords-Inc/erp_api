-- §7.1 rows 13-14: recompute the affected SO line's status whenever an item release
-- line is created or its released_qty changes. sales_order_details_id is a direct FK.
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_item_release]
ON [dbo].[tbl_inv_item_release_details]
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
