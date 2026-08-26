-- §7.1 rows 16-17 (FAILED/RETURN, DELIVERED): recompute every SO line on the delivery
-- receipt a logistics route leg is tracking, whenever departed_at/arrived_at/
-- returned_at change. delivery_receipt_doc is a doc-number string, not a numeric FK
-- (doc_no is uniquely indexed, so this is reliable even though it isn't a clean id
-- join).
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_logistics_route]
ON [dbo].[tbl_dispatching_logistics_route]
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @id BIGINT;
    DECLARE cur CURSOR LOCAL FAST_FORWARD FOR
        SELECT DISTINCT dri.sales_order_details_id
        FROM inserted i
        INNER JOIN tbl_dispatching_delivery_receipt dr
            ON dr.doc_no = TRY_CAST(i.delivery_receipt_doc AS BIGINT)
        INNER JOIN tbl_dispatching_delivery_receipt_items dri
            ON dri.delivery_receipt_id = dr.id
        WHERE dri.sales_order_details_id IS NOT NULL;

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
