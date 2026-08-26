-- §7.1 rows 5-10: recompute the affected SO line's status whenever a Job Order row
-- changes (created, assigned, accepted, item request raised, marked complete, WH
-- acknowledged). order_details_id is a direct FK - no comma-list parsing needed.
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_job_order]
ON [dbo].[tbl_trans_job_order]
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @id BIGINT;
    DECLARE cur CURSOR LOCAL FAST_FORWARD FOR
        SELECT DISTINCT order_details_id FROM inserted WHERE order_details_id IS NOT NULL;

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
