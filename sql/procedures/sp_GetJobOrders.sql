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
--
-- MATERIALS rewritten 2026-09-05 (user-reported bug: "I think the stock here is
-- just hardcoded"). jo.materials is a real column, but nothing anywhere in the API
-- ever writes to it (confirmed by search) - it sat NULL forever, and the client's
-- own fallback (JobOrderPage.ApplyIncompleteToMaterials) just displays the literal
-- string "INCOMPLETE" whenever it's blank. So every job order showed INCOMPLETE,
-- permanently, regardless of actual stock - there was no live computation behind
-- it at all, in either direction.
--
-- Computed live instead (user decision - never write it back, so it can never go
-- stale after an RR receipt or an approval): COMPLETE only if EVERY component
-- under this row's resolved BOM has enough real stock, using the same
-- physical-minus-reserved math sp_GetComponents now uses (see that file's header
-- for the full reasoning), including the same carve-out for this job's own SO's
-- own approved reservation counting as covering the need rather than competing
-- demand.
CREATE OR ALTER PROCEDURE [dbo].[sp_GetJobOrders] @UserId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT ISNULL(jo.id, 0) AS id,
    resolved.bom_id AS bom_id,
    so.order_id AS so_id,
    sod.order_details_id AS order_details_id,
    so.date AS date,
    so.doc AS sales_order,
    matcheck.materials AS materials,
    sod.qty AS quantity,
    jo.due AS due,
    jo.engr_id,
    jo.a_engr AS a_engr,
    -- jo.item_rqst was ALSO aliased "item_rqst" here, alongside ir.doc_no below -
    -- two result columns with the same name, so which one survived JSON marshalling
    -- was not deterministic. The live request number (ir.doc_no) is the one the grid
    -- needs; jo.item_rqst is a legacy free-text column nothing writes.
    jo.status AS status,
    sod.status AS so_item_status,
    itn.name AS general_name,
    jo.serial_no AS serial_no,
    jo.report AS report,
    jo.report_base AS report_base,
    i.item_model AS type,
    ias.long_description AS item_desc,
    req.doc_no AS item_rqst,
    req.id AS ir_id
FROM dbo.tbl_trans_sales_order AS so
    INNER JOIN dbo.tbl_trans_sales_order_details AS sod ON so.order_id = sod.based_id
    INNER JOIN dbo.tbl_setup_item AS i ON sod.item_id = i.id
    INNER JOIN dbo.tbl_setup_item_name AS itn ON i.item_name_id = itn.id
    -- Resolves this line's BOM from its item - see the bom_id note in the header.
    -- LEFT, so a line whose item has no BOM still appears (with bom_id 0).
    LEFT JOIN dbo.tbl_setup_item_bom AS itembom ON itembom.item_id = sod.item_id
    INNER JOIN tbl_setup_item_additional_specs ias ON ias.based_id = i.id
    LEFT JOIN dbo.tbl_trans_job_order AS jo ON sod.order_details_id = jo.order_details_id
    -- ITEM REQUEST # rewritten 2026-09-05. This joined tbl_inv_item_request /
    -- tbl_inv_item_request_details - the ORIGINAL tables, which the module was moved
    -- off at some point in favour of tbl_inv_item_request2 / _details2. The originals
    -- still exist but hold zero rows, so this column could only ever come back NULL no
    -- matter how many item requests were raised. The Go side had already moved
    -- (models/inventory_models/inventory_item_request_model.go maps the "2" tables), as
    -- had sp_GetSalesOrderDetailsItemReq, vw_get_item_request_details and
    -- vw_get_sales_order_item_req_doc; this procedure was left behind.
    --
    -- The "2" schema also renamed the keys: the child FK is item_request_id (was ir_id)
    -- and the SO line is sales_order_details_id (was sod_id), while the parent now
    -- carries sales_order_id directly.
    --
    -- Matched per SO LINE rather than per sales order. The old join keyed on
    -- so.order_id, so any request against the order attached itself to every job order
    -- row on it; ITEM REQUEST # is meant to show the request covering THIS item.
    -- OUTER APPLY TOP 1 rather than a LEFT JOIN so that several requests against one
    -- line (a partial request followed by another) cannot fan out and duplicate job
    -- order rows - the newest wins.
    OUTER APPLY (
        SELECT TOP 1 ir2.id, ir2.doc_no
        FROM dbo.tbl_inv_item_request_details2 AS ird2
            INNER JOIN dbo.tbl_inv_item_request2 AS ir2
                ON ir2.id = ird2.item_request_id
        WHERE ird2.sales_order_details_id = sod.order_details_id
        ORDER BY ir2.doc_no DESC
    ) req
    CROSS APPLY (
        SELECT COALESCE(NULLIF(sod.bom_id, 0), itembom.id, 0) AS bom_id
    ) resolved
    -- Real MATERIALS status - see header. NOT EXISTS a short component -> COMPLETE;
    -- a BOM with no detail rows at all vacuously has none, so it reads COMPLETE too
    -- (nothing needed, nothing missing) rather than crashing on an empty result.
    CROSS APPLY (
        SELECT CASE WHEN EXISTS (
            SELECT 1
            FROM dbo.tbl_setup_item_bom_details AS bod
                OUTER APPLY (
                    SELECT SUM(stock_qty) AS physical
                    FROM dbo.tbl_inv_item_stocks
                    WHERE item_id = bod.item_id
                ) phys
                OUTER APPLY (
                    SELECT SUM(qty) AS qty
                    FROM dbo.tbl_inv_stock_reservations
                    WHERE item_id = bod.item_id AND quotation_id = so.quotation_id AND status = 'Approved'
                ) this_job
                OUTER APPLY (
                    SELECT SUM(qty) AS reserved
                    FROM dbo.tbl_inv_stock_reservations
                    WHERE item_id = bod.item_id AND status <> 'Rejected'
                        AND NOT (quotation_id = so.quotation_id AND status = 'Approved')
                ) other_resv
            WHERE bod.item_bom_id = resolved.bom_id
                AND bod.bom_qty > (ISNULL(this_job.qty, 0) + (ISNULL(phys.physical, 0) - ISNULL(other_resv.reserved, 0)))
        ) THEN 'INCOMPLETE' ELSE 'COMPLETE' END AS materials
    ) matcheck
-- Approval gate added 2026-09-05 (user requirement: "Job order will only display if
-- the Sales order is approved").
--
-- This clause used to read "so.approved_by_id = @UserId", which is an OWNERSHIP test,
-- not an approval one - and it let unapproved orders straight through, because an
-- unapproved sales order carries approved_by_id = 0. Any caller whose user id resolved
-- to 0 therefore saw every unapproved order: SO#0007 (approved_by_id 0, status '-')
-- was sitting in the ONGOING tab despite never having been approved, which §14.7
-- forbids - an SO must not reach another department before approval.
--
-- The per-user restriction is dropped as well (user decision). §6.3: once approved an
-- SO "becomes visible in every department's SO view", and §6.1 has engineers taking
-- work from a shared pool that the head engineer assigns from - neither is a
-- "only orders I personally approved" view. The PENDING / ONGOING / FINISHED tabs
-- already segment by assignment client-side (a_engr / due / status), which is the
-- split that actually matters here.
--
-- @UserId is deliberately still declared: the endpoint is
-- /engineering/job_order/:user_id and every caller passes it. Removing the parameter
-- would break them for no gain, so it stays unused rather than silently changing the
-- route's shape.
--
-- status is compared through ISNULL so a NULL never evaluates to UNKNOWN and quietly
-- drops an otherwise valid row. Live values today are 'ACTIVE' and '-'.
WHERE so.approved_by_id > 0
    AND ISNULL(so.status, '') <> 'CANCELLED'   -- §14.8: a cancelled SO permits no further progress
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
