-- STEP 01 of 07 - adds the remarks column 07 writes into - run before 07 or the remarks are only counted
-- Run the numbered scripts in this folder in order, each on its own: 01, 02 ... 07.
-- Run before: 02_add_project_content_flow_head_id_2026-09-23.sql
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- A remarks column on the BOM component line
-- ==========================================
-- tbl_setup_item_bom_details had nowhere to record what a component is FOR on its BOM -
-- 'Power', 'Control', 'Wye', 'Main, Delta', 'alternative models'. The controller workbook
-- carries one on 39,180 of its 130,372 lines, and 07_load_controllers_bom_*.sql writes them
-- only if this column exists (it checks with COL_LENGTH and reports them instead when it
-- does not). So run this BEFORE that script, or run that script again afterwards - it fills
-- the remarks on BOMs it has already loaded.
--
-- Starting the API against the database does exactly the same thing: Remarks is declared on
-- SetupItemBomDetailsContent and migrateBomBoq() is live in MigrateAll(), so GORM adds the
-- column and its audit twin on start-up. This script is here for a database that has to be
-- migrated before the API is run against it, and for one where the API is not restarted.
--
-- Idempotent: adds each column only if it is missing, changes nothing else.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

BEGIN TRANSACTION;

IF COL_LENGTH(N'tbl_setup_item_bom_details', N'remarks') IS NULL
BEGIN
    ALTER TABLE tbl_setup_item_bom_details ADD remarks nvarchar(400) NULL;
    PRINT 'added tbl_setup_item_bom_details.remarks';
END
ELSE
    PRINT 'tbl_setup_item_bom_details.remarks already there';

-- The audit twin, so a line's remark is kept in its history like every other field.
IF OBJECT_ID(N'z_tbl_setup_item_bom_details_at') IS NOT NULL
   AND COL_LENGTH(N'z_tbl_setup_item_bom_details_at', N'remarks') IS NULL
BEGIN
    ALTER TABLE z_tbl_setup_item_bom_details_at ADD remarks nvarchar(400) NULL;
    PRINT 'added z_tbl_setup_item_bom_details_at.remarks';
END

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed - now refresh vw_item_bom_list (sql/views/vw_item_bom_list.sql), which
       selects the new column, by restarting the API or running that file';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'rolled back - set @commit = 1 to write';
END
