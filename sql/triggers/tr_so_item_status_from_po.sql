-- §7.1 row 3 (WAITING FOR DELIVERY): recompute every SO line consolidated onto a PO
-- detail row whenever that row is created or changed. order_detail_ids is a
-- comma-joined string (one PO line can consolidate the same item across several SOs,
-- CLAUDE.md invariant #7) - matched by LIKE against tbl_trans_sales_order_details
-- directly rather than split, since this DB's compatibility level (110) has no
-- STRING_SPLIT.
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_po]
ON [dbo].[tbl_purchasing_purchase_order_details]
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @id BIGINT;
    DECLARE cur CURSOR LOCAL FAST_FORWARD FOR
        SELECT DISTINCT sod.order_details_id
        FROM inserted i
        INNER JOIN tbl_trans_sales_order_details sod
            ON ',' + ISNULL(i.order_detail_ids, '') + ','
               LIKE '%,' + CAST(sod.order_details_id AS NVARCHAR(20)) + ',%';

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
