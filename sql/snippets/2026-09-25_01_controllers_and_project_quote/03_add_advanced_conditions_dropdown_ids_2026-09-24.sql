-- STEP 03 of 07 - the six Advanced Conditions dropdown ids
-- Run the numbered scripts in this folder in order, each on its own: 01, 02 ... 07.
-- Run after:  02_add_project_content_flow_head_id_2026-09-23.sql
-- Run before: 04_fix_template_child_level_zero_2026-09-24.sql
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- The six Advanced Conditions dropdowns, stored as IDs
-- ====================================================
-- BRAND, DRIVER TYPE, MOTOR ENCLOSURE, MOTOR MANUFACTURER, LIQUID TYPE and CONTROLLER
-- MANUFACTURER were free-text boxes on a project quotation's Advanced Conditions panel.
-- They are dropdowns now, and what is stored is the id, not the words (user decision
-- 2026-09-24 - the same rule as flow_id / head_id on the content row).
--
--   pump_brand_id               -> tbl_setup_item_brand, the list Item Entry uses. A brand
--                                  renamed in setup then reads correctly on every
--                                  quotation already saved.
--   driver_type_id              1 Electric Motor, 2 Engine
--   motor_enclosure_id          1 NEMA 2, 2 NEMA 3, 3 NEMA 4, 4 NEMA 12
--   motor_manufacturer_id       1 Teco Motor, 2 Franklin Motor, 3 Calpeda Motor
--   liquid_type_id              1 Clean, 2 Sewage, 3 Seawater
--   controller_manufacturer_id  1 Eaton, 2 Marathon, 3 WEG
--
-- Those five lists live in the client for now; storing the id is what lets any of them
-- become a Setup list later without rewriting a single stored quotation. THE IDS ARE
-- FIXED: append, never renumber. 0 means "-- Select --", which is what an unanswered
-- dropdown stores - not the first value.
--
-- The existing nvarchar columns (pump_brand, driver_type, ...) are KEPT and left alone.
-- They are empty on every row that exists, so nothing is lost by ignoring them, and
-- dropping a column is not something this needs to do.
--
-- Starting the API adds these columns too - all six are on
-- SalesProjectAdvancedConditionsContent and migrateSalesProject() is live in MigrateAll().
-- This is for a database migrated before the API is first run against it.
--
-- Idempotent: adds each column only if it is missing.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

DECLARE @table sysname = N'tbl_trans_sales_project_content_advanced_configuration';
DECLARE @audit sysname = N'tbl_trans_sales_project_content_advanced_configuration_at';

IF OBJECT_ID('tempdb..#cols') IS NOT NULL DROP TABLE #cols;
CREATE TABLE #cols (name sysname);
INSERT INTO #cols (name) VALUES
    (N'pump_brand_id'), (N'driver_type_id'), (N'motor_enclosure_id'),
    (N'motor_manufacturer_id'), (N'liquid_type_id'), (N'controller_manufacturer_id');

BEGIN TRANSACTION;

DECLARE @name sysname, @sql nvarchar(max), @target sysname;

DECLARE targets CURSOR LOCAL FAST_FORWARD FOR
    SELECT t.tbl, c.name
      FROM (SELECT @table AS tbl UNION ALL SELECT @audit) t
      CROSS JOIN #cols c
     WHERE OBJECT_ID(t.tbl) IS NOT NULL;

OPEN targets;
FETCH NEXT FROM targets INTO @target, @name;
WHILE @@FETCH_STATUS = 0
BEGIN
    IF COL_LENGTH(@target, @name) IS NULL
    BEGIN
        SET @sql = N'ALTER TABLE ' + QUOTENAME(@target) + N' ADD ' + QUOTENAME(@name) + N' bigint NULL;';
        EXEC sp_executesql @sql;
        PRINT 'added ' + @target + '.' + @name;
    END
    ELSE
        PRINT @target + '.' + @name + ' already there';

    FETCH NEXT FROM targets INTO @target, @name;
END
CLOSE targets;
DEALLOCATE targets;

-- Nothing is backfilled. The text columns these replace are empty on every existing row,
-- so there is nothing to convert - and 0/NULL is "-- Select --", which is the truth for a
-- question nobody has answered.
EXEC sp_executesql N'
SELECT ''rows with nothing chosen yet'' AS state, COUNT(*) AS rows_
  FROM tbl_trans_sales_project_content_advanced_configuration
 WHERE ISNULL(pump_brand_id, 0) = 0;';

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

IF OBJECT_ID('tempdb..#cols') IS NOT NULL DROP TABLE #cols;
