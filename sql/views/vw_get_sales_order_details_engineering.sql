-- Sales Order lines offered to Engineering when raising a Job Order.
--
-- Only BOM HEADS appear here, not every line of an expanded BOM (user decision,
-- 2026-09-03). For a quotation shaped like:
--
--   1        COMMON PACKAGE            <- head
--   1.1      DISCHARGE COMMON HEADER   <- head (a child of 1, but heads its own BOM)
--   1.1.1    PRESSURE GAUGE
--   1.1.2    DISCHARGE RUBBER BELOW
--   1.2      SUCTION COMMON HEADER
--
-- only 1 and 1.1 are job-order material - the rest are components consumed by
-- assembling them.
--
-- The test is "does this line's item own a BOM", i.e. does tbl_setup_item_bom have
-- a row for it. Two alternatives were checked and rejected:
--
--   * bom_id alone cannot work - a head and its children carry the SAME bom_id
--     (1.1 through 1.1.5 are all bom_id = 3). It records which BOM a row came
--     from, not whether the row heads one.
--   * bom_id + item_id does work, but needs bom_id carried onto the sales order
--     (it isn't today - 0 of 27 rows have it) AND populated on project items
--     (never set - 0 of 72 rows), so it would have taken a schema change plus two
--     conversion fixes and still left project quotes broken until both landed.
--
-- Keying on item_id alone gives an identical result on quick quotes and also works
-- on project items, which carry no bom_id at all - verified against live data
-- before this filter was written. It needs nothing added to the sales order, since
-- item_id is already there.
--
-- A line whose item has no BOM is excluded on purpose: there is nothing to
-- fabricate for a plain traded item, so it has no place on a Job Order.
-- STOCK fixed 2026-09-05 (user question: "the stock there is accurate based on our
-- stock table?" - it was not).
--
-- This column read "sod.qty AS stock", i.e. the ORDERED QUANTITY aliased as stock,
-- and req_qty was the same expression - so both columns always showed the identical
-- number and neither had any connection to inventory. The view joined no stock table
-- at all. Live example before the fix: SO#0007 reported stock 10 for items 1252 and
-- 1263, whose real physical stock is 0 and 1.
--
-- Now the real figure, using the SAME formula sp_GetComponents and sp_GetJobOrders'
-- matcheck already use, so the three can never disagree:
--
--     available = this SO's own approved reservation
--               + (physical stock - reservations held by anyone else)
--
-- §8.11: "TOTAL STOCK (tracker) = Σ zone units, EXCLUDING reserved", and §14.18
-- forbids counting reserved units inside it. The carve-out for this order's own
-- approved reservation is §10.4.4 - a hold raised for this order exists precisely to
-- fulfil it, so counting it as competing demand would make a fully-reserved line read
-- as short. §8.12 applies the same correction on the purchasing side.
--
-- Clamped at 0: only Item Release may drive stock negative, and only against a vehicle
-- zone (§10.5, invariant 4). A negative here would be a display artefact, not a fact.
CREATE OR ALTER VIEW [dbo].[vw_get_sales_order_details_engineering] AS
SELECT sod.order_details_id AS id,
    sod.based_id AS so_id,
    sod.item_code AS item_code,
    sod.item_description AS item_desc,
    CASE
        WHEN ISNULL(this_job.qty, 0) + (ISNULL(phys.physical, 0) - ISNULL(other_resv.reserved, 0)) < 0
        THEN 0
        ELSE ISNULL(this_job.qty, 0) + (ISNULL(phys.physical, 0) - ISNULL(other_resv.reserved, 0))
    END AS stock,
    sod.qty AS req_qty,
    sod.delivery_preference AS remark,
    sod.status AS status
FROM dbo.tbl_trans_sales_order_details AS sod
    LEFT JOIN dbo.tbl_trans_sales_order AS so ON sod.based_id = so.order_id
    OUTER APPLY (
        SELECT SUM(st.stock_qty) AS physical
        FROM dbo.tbl_inv_item_stocks AS st
        WHERE st.item_id = sod.item_id
    ) phys
    OUTER APPLY (
        SELECT SUM(r.qty) AS qty
        FROM dbo.tbl_inv_stock_reservations AS r
        WHERE r.item_id = sod.item_id
            AND r.quotation_id = so.quotation_id
            AND r.status = 'Approved'
    ) this_job
    OUTER APPLY (
        SELECT SUM(r.qty) AS reserved
        FROM dbo.tbl_inv_stock_reservations AS r
        WHERE r.item_id = sod.item_id
            AND r.status <> 'Rejected'
            AND NOT (r.quotation_id = so.quotation_id AND r.status = 'Approved')
    ) other_resv
WHERE EXISTS (
        SELECT 1
        FROM dbo.tbl_setup_item_bom AS b
        WHERE b.item_id = sod.item_id
    )
