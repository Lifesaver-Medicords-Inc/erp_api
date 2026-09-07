-- Components of one BOM, for the Job Order screen's MATERIALS popup
-- (MaterialsForm -> GET /engineering/job_order/components/:bom_id?so_id=).
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
-- STOCK rewritten again 2026-09-05 (user-reported bug: "the stock here is just
-- hardcoded"). It was "bod.bom_qty AS stock" - a copy of the required quantity, so
-- a component could never show short. Now real, using the same physical-minus-
-- reserved math the Sales stock-check button already uses
-- (ItemStockService.GetAvailableStock):
--
--     available = SUM(tbl_inv_item_stocks.stock_qty)      -- physical on hand,
--                                                             already reflects
--                                                             every RR receipt
--               - SUM(tbl_inv_stock_reservations.qty)      -- every OTHER active
--                 WHERE status <> 'Rejected'                  reservation (Pending
--                                                             or Approved both
--                                                             hold stock back)
--
-- With one addition that generic formula doesn't have: this job's OWN sales
-- order's own APPROVED reservation for the component counts as covering the
-- need, not as competing demand (user decision - "if that item have stock in
-- sales quotation requested and approved by the inventory"; matches CLAUDE.md
-- invariant #5, "An SO fulfils from its own reserved units first"). So it's
-- carved out of "other" and added back on top instead:
--
--     stock = this_job's own approved reservation for this item
--           + (physical - every OTHER active reservation for this item)
--
-- @SoId is optional (defaults 0, matching @BomId's own style elsewhere in this
-- file) so a caller that doesn't have it yet still gets a real, if slightly more
-- conservative (no carve-out), stock figure rather than an error.
CREATE OR ALTER PROCEDURE [dbo].[sp_GetComponents] @BomId int, @SoId int = 0 AS BEGIN
SET NOCOUNT ON;
BEGIN TRY

DECLARE @QuotationId INT = (SELECT quotation_id FROM dbo.tbl_trans_sales_order WHERE order_id = @SoId);

SELECT bod.id AS id,
    itn.name AS name,
    bod.bom_qty AS quantity,
    ISNULL(this_job.qty, 0) + (ISNULL(phys.physical, 0) - ISNULL(other_resv.reserved, 0)) AS stock
FROM dbo.tbl_setup_item_bom_details AS bod
    INNER JOIN dbo.tbl_setup_item AS i ON bod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
    OUTER APPLY (
        SELECT SUM(stock_qty) AS physical
        FROM dbo.tbl_inv_item_stocks
        WHERE item_id = bod.item_id
    ) phys
    -- This job's own SO's own approved reservation - covers the need.
    OUTER APPLY (
        SELECT SUM(qty) AS qty
        FROM dbo.tbl_inv_stock_reservations
        WHERE item_id = bod.item_id AND quotation_id = @QuotationId AND status = 'Approved'
    ) this_job
    -- Every OTHER active reservation - held back from what this job can draw on.
    -- Excludes the exact rows this_job already counted, so nothing is both added
    -- and subtracted.
    OUTER APPLY (
        SELECT SUM(qty) AS reserved
        FROM dbo.tbl_inv_stock_reservations
        WHERE item_id = bod.item_id AND status <> 'Rejected'
            AND NOT (quotation_id = @QuotationId AND status = 'Approved')
    ) other_resv
WHERE bod.item_bom_id = @BomId;

END TRY BEGIN CATCH THROW;
END CATCH
END;
