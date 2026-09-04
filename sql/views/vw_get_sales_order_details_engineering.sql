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
CREATE OR ALTER VIEW [dbo].[vw_get_sales_order_details_engineering] AS
SELECT sod.order_details_id AS id,
    sod.based_id AS so_id,
    sod.item_code AS item_code,
    sod.item_description AS item_desc,
    sod.qty AS stock,
    sod.qty AS req_qty,
    sod.delivery_preference AS remark,
    sod.status AS status
FROM dbo.tbl_trans_sales_order_details AS sod
    LEFT JOIN dbo.tbl_trans_sales_order AS so ON sod.based_id = so.order_id
WHERE EXISTS (
        SELECT 1
        FROM dbo.tbl_setup_item_bom AS b
        WHERE b.item_id = sod.item_id
    )
