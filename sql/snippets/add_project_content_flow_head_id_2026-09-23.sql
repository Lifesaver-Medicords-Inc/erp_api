-- FLOW and HEAD unit columns on a project quotation's item set
-- ============================================================
-- The two little dropdowns beside FLOW and HEAD on a project quotation (ItemSetUC) had
-- nowhere to be saved: tbl_trans_sales_project_content carries flow and head, the figures,
-- and nothing for the unit each was quoted in. The controls were not even read - whatever
-- was picked went nowhere.
--
-- The unit is stored as an ID, not as its text. The lists stay hard-coded in the client for
-- now (user decision, 2026-09-23), and the columns are named after the two controls,
-- cmb_flow and cmb_head - Helpers.GetControlsValues turns a combo tagged DYNAMIC into
-- <its name without cmb_>_id, which is the key the quotation sends:
--
--     flow_id   1 = gpm lpm      head_id   1 = psi
--               2 = m^2/hr                2 = ft
--               3 = lps                   3 = m
--
-- Storing the id is what makes either list a Setup table later without touching a single
-- saved quotation. 0 (or NULL) means nothing was chosen and reads as the first entry, id 1.
--
-- Starting the API does the same thing - both fields are on SalesProjectContentContent and
-- migrateSalesProject() is live in MigrateAll() - so this is for a database that has to be
-- migrated before the API is ever run against it.
--
-- Idempotent: adds each column only if it is missing, changes nothing else.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

BEGIN TRANSACTION;

IF COL_LENGTH(N'tbl_trans_sales_project_content', N'flow_id') IS NULL
BEGIN
    ALTER TABLE tbl_trans_sales_project_content ADD flow_id bigint NULL;
    PRINT 'added tbl_trans_sales_project_content.flow_id';
END
ELSE
    PRINT 'tbl_trans_sales_project_content.flow_id already there';

IF COL_LENGTH(N'tbl_trans_sales_project_content', N'head_id') IS NULL
BEGIN
    ALTER TABLE tbl_trans_sales_project_content ADD head_id bigint NULL;
    PRINT 'added tbl_trans_sales_project_content.head_id';
END
ELSE
    PRINT 'tbl_trans_sales_project_content.head_id already there';

-- The audit twin, so a change of unit is kept in the row's history like every other field.
IF OBJECT_ID(N'z_tbl_trans_sales_project_content_at') IS NOT NULL
BEGIN
    IF COL_LENGTH(N'z_tbl_trans_sales_project_content_at', N'flow_id') IS NULL
    BEGIN
        ALTER TABLE z_tbl_trans_sales_project_content_at ADD flow_id bigint NULL;
        PRINT 'added z_tbl_trans_sales_project_content_at.flow_id';
    END

    IF COL_LENGTH(N'z_tbl_trans_sales_project_content_at', N'head_id') IS NULL
    BEGIN
        ALTER TABLE z_tbl_trans_sales_project_content_at ADD head_id bigint NULL;
        PRINT 'added z_tbl_trans_sales_project_content_at.head_id';
    END
END

-- Nothing is backfilled. An existing item set has no recorded unit, which is the truth -
-- it was never asked for - and 0/NULL is what the client shows as the first entry.
--
-- Through sp_executesql because the whole script is one batch: a plain SELECT naming a
-- column added higher up is resolved when the batch is parsed, before the ALTER has run,
-- and fails with "Invalid column name".
EXEC sp_executesql N'
SELECT ''item sets with no unit recorded'' AS state, COUNT(*) AS rows_
  FROM tbl_trans_sales_project_content
 WHERE ISNULL(flow_id, 0) = 0 OR ISNULL(head_id, 0) = 0;';

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END
