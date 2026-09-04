-- Job Order grid (GET /engineering/job_order/:user_id).
--
-- bom_id changed 2026-09-03: it used to be plain sod.bom_id, which is 0 on every
-- sales order line ever written - the quote -> SO conversion never copies it. That
-- 0 reached JobOrderPage's MATERIALS cell, which explicitly refuses to open with
-- "No BOM ID found for this record", so the materials list was unreachable for
-- every job order regardless of whether its item actually had a BOM.
--
-- Resolved from the item instead, which is data that already exists, so old sales
-- orders work with no backfill and no schema change. The stored value still wins
-- when it is populated, so if the conversion is fixed later to carry bom_id
-- through, nothing here needs revisiting. Falls back to 0 when the item genuinely
-- has no BOM - the client's existing "No BOM ID" guard is then correct, because
-- there is nothing to fabricate.
--
-- Safe against fan-out: tbl_setup_item_bom holds at most one row per item_id
-- (verified). If that ever stops being true this join would duplicate job order
-- rows, so it would need an explicit "pick one" rule at that point.
CREATE OR ALTER PROCEDURE [dbo].[sp_GetJobOrders] @UserId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT ISNULL(jo.id, 0) AS id,
    COALESCE(NULLIF(sod.bom_id, 0), itembom.id, 0) AS bom_id,
    so.order_id AS so_id,
    sod.order_details_id AS order_details_id,
    so.date AS date,
    so.doc AS sales_order,
    jo.materials AS materials,
    sod.qty AS quantity,
    jo.due AS due,
    jo.engr_id,
    jo.a_engr AS a_engr,
    jo.item_rqst AS item_rqst,
    jo.status AS status,
    sod.status AS so_item_status,
    itn.name AS general_name,
    jo.serial_no AS serial_no,
    jo.report AS report,
    jo.report_base AS report_base,
    i.item_model AS type,
    ias.long_description AS item_desc,
    ir.doc_no AS item_rqst,
    ir.id AS ir_id
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN dbo.tbl_trans_sales_order_details AS sod ON so.order_id = sod.based_id
    INNER JOIN dbo.tbl_setup_item AS i ON sod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
    -- Resolves this line's BOM from its item - see the bom_id note in the header.
    -- LEFT, so a line whose item has no BOM still appears (with bom_id 0).
    LEFT JOIN dbo.tbl_setup_item_bom AS itembom ON itembom.item_id = sod.item_id
    INNER JOIN tbl_setup_item_additional_specs ias ON ias.based_id = i.id
    LEFT JOIN dbo.tbl_trans_job_order AS jo ON sod.order_details_id = jo.order_details_id
    LEFT JOIN dbo.tbl_inv_item_request_details AS ird ON so.order_id = ird.so_id
    LEFT JOIN dbo.tbl_inv_item_request AS ir ON ird.ir_id = ir.id
WHERE so.approved_by_id = @UserId
    -- Only BOM heads are job-order material (user decision, 2026-09-03). For a
    -- quotation shaped 1 / 1.1 / 1.1.1 / 1.2, only 1 and 1.1 belong here - the rest
    -- are components consumed by assembling them, and listing them produced rows
    -- whose MATERIALS cell could only ever answer "No BOM ID found for this
    -- record", because a component has no BOM of its own.
    --
    -- Keyed on the item owning a BOM, the same test vw_get_sales_order_details_engineering
    -- uses. Deliberately NOT sod.bom_id: a head and its children carry the same
    -- bom_id, and on the sales order it is 0 everywhere anyway.
    --
    -- Applied strictly (user decision, 2026-09-03): a line whose item has no BOM is
    -- excluded even when a job order row already exists for it. Two such rows do
    -- exist (ids 1 and 2, PUMP, created before this rule) and are now hidden here.
    -- They still exist in tbl_trans_job_order - this filters the grid, it does not
    -- delete anything - but nothing in this screen reaches them any more.
    AND EXISTS (
        SELECT 1
        FROM dbo.tbl_setup_item_bom AS b
        WHERE b.item_id = sod.item_id
    );
END TRY BEGIN CATCH THROW;
END CATCH
END;