-- §7.1 row 4 (RR made against PO -> IN STOCK): recompute every SO line consolidated
-- onto the PO detail row a Receiving Report line was received against. Same
-- comma-list LIKE-join as tr_so_item_status_from_po, one hop further back
-- (RR details -> PO details -> SO details).
CREATE OR ALTER TRIGGER [dbo].[tr_so_item_status_from_rr]
ON [dbo].[tbl_inv_receiving_report_details]
AFTER INSERT, UPDATE
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @id BIGINT;
    DECLARE cur CURSOR LOCAL FAST_FORWARD FOR
        SELECT DISTINCT sod.order_details_id
        FROM inserted i
        INNER JOIN tbl_purchasing_purchase_order_details pod ON pod.id = i.purchase_order_details_id
        INNER JOIN tbl_trans_sales_order_details sod
            ON ',' + ISNULL(pod.order_detail_ids, '') + ','
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
