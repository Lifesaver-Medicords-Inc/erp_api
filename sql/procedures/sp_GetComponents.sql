-- Components of one BOM, for the Job Order screen's MATERIALS popup
-- (MaterialsForm -> GET /engineering/job_order/components/:bom_id).
--
-- Rewritten 2026-09-03. The previous version returned NOTHING, ever:
--
--     FROM tbl_trans_sales_order_details AS so
--     INNER JOIN tbl_setup_item_bom AS bom ON so.bom_id = bom.id
--     ...
--     WHERE so.bom_id = @BomId;
--
-- It reached the BOM by way of tbl_trans_sales_order_details.bom_id, which is 0 on
-- every sales order line ever written - the quote -> SO conversion never copies it
-- (Orders.cs stamps a flat counter into "numbering" and leaves bom_id alone). So
-- the INNER JOIN matched no rows and the materials list was permanently empty.
-- Confirmed by running it against BOM 3, which has 5 components, and getting none.
--
-- The sales order join was also pointless on its own terms: this proc answers
-- "what is in BOM @BomId", a question about setup data that has nothing to do with
-- any particular sales order. Worse, it fanned out - had bom_id ever been
-- populated, every component would have been returned once per matching SO line.
--
-- Now it reads the BOM directly, so it works for existing sales orders with no
-- backfill and no schema change.
CREATE OR ALTER PROCEDURE [dbo].[sp_GetComponents] @BomId int AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT bod.id AS id,
    itn.name AS name,
    bod.bom_qty AS quantity,
    bod.bom_qty AS stock
FROM dbo.tbl_setup_item_bom_details AS bod
    INNER JOIN dbo.tbl_setup_item AS i ON bod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
WHERE bod.item_bom_id = @BomId;
END TRY BEGIN CATCH THROW;
END CATCH
END;
