IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_order_item_release' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_order_item_release] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_order_item_release]
AS
SELECT
    sod.order_details_id AS sales_order_details_id,
    sod.based_id AS sales_order_id,
    so.doc as ref_doc_no,
    sod.item_id,
    -- Fall back to the item's model when the SO line carries no description of its own.
    --
    -- sod.item_description is free text typed on the sales order line, and most lines are
    -- saved without it - so the Item Release grid (and the printed IREL, which uses the same
    -- column) showed an empty description for nearly every row, while the SAME item read
    -- correctly on the one line where somebody had typed something. Identifying an item by
    -- code alone is not enough for the warehouse to pick against.
    -- (user-reported 2026-09-05)
    --
    -- The SO line still wins wherever it has text: it describes what was actually sold,
    -- which may deliberately differ from the catalogue entry. This only fills the blank.
    COALESCE(NULLIF(LTRIM(RTRIM(sod.item_description)), ''), i.item_model) AS item_description,
    sod.qty AS required_qty,
    i.unit_of_measure_id as required_uom_id,
    uom.name as required_uom,
    sod.delivery_preference,
    sod.item_code,
	ird.released_qty,
	ird.released_uom_id as released_uom
FROM tbl_trans_sales_order_details sod
LEFT JOIN tbl_trans_sales_order so
	ON sod.based_id = so.order_id
LEFT JOIN tbl_setup_item i
    ON sod.item_id = i.id
LEFT JOIN tbl_setup_item_unit_measurement uom
	ON i.unit_of_measure_id = uom.id
-- Aggregated before the join, not joined raw: tbl_inv_item_release_details has one
-- row per (sales_order_id, item_id) PER Item Release document, and partial releases
-- are explicitly legitimate (spec: "Partial releases produce multiple PAs") - a raw
-- join fanned out one duplicate SO detail row per PRIOR Item Release against the
-- same SO, so the 2nd (and 3rd, 4th...) Item Release ever created against an SO
-- showed every item multiplied by however many releases already existed. This
-- restores the join to one row per (sales_order_id, item_id) regardless of history.
LEFT JOIN (
	SELECT sales_order_id, item_id,
		SUM(released_qty) AS released_qty,
		MAX(released_uom_id) AS released_uom_id
	FROM tbl_inv_item_release_details
	GROUP BY sales_order_id, item_id
) ird
	ON ird.sales_order_id = sod.based_id
	AND ird.item_id = sod.item_id
