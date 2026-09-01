/*
 * Clears every transactional table while preserving Setup/master data -
 * Item Entry, Chart of Accounts, BPI, Access Modules/Positions, Company Setup,
 * warehouse structure, and the Project Quotation default template.
 *
 * Built for the local deploy rehearsal (see the deployment plan memory /
 * D:\Claude\plans\glistening-humming-glade.md) but written to be reusable
 * for the real server move too: restore the live DB in full first (so every
 * ALTER VIEW/PROCEDURE succeeds - see that plan for why a truly empty DB
 * can't start today), then run this against the restored copy to strip out
 * accumulated test transactions before real users touch it.
 *
 * NEVER run this against a database anyone is actively relying on for real
 * data - it deletes every row from every table not explicitly listed below.
 *
 * Safe to re-run: FK constraints and the one known trigger
 * (tr_inv_item_stocks_ledger) are disabled for the duration and restored
 * (WITH CHECK, re-validating) at the end.
 */

SET NOCOUNT ON;

-- Tables whose data is Setup/master data and must survive - everything else
-- gets emptied. z_-prefixed audit ("_at") tables are deliberately NEVER
-- listed here: they're a log of past changes, not something the app reads
-- back, and are exactly the kind of test-activity history this reset is
-- for removing.
IF OBJECT_ID('tempdb..#KeepTables') IS NOT NULL DROP TABLE #KeepTables;
CREATE TABLE #KeepTables (name NVARCHAR(128) PRIMARY KEY);
INSERT INTO #KeepTables (name) VALUES
    ('tbl_access_modules'),
    ('tbl_admin_position'),
    ('tbl_admin_position_access'),
    ('tbl_bpi'),
    ('tbl_bpi_accreditation'),
    ('tbl_bpi_address'),
    ('tbl_bpi_branch_industries'),
    ('tbl_bpi_contacts'),
    ('tbl_bpi_entity'),
    ('tbl_bpi_finance'),
    ('tbl_bpi_general'),
    ('tbl_bpi_history'),
    ('tbl_bpi_industries'),
    ('tbl_bpi_items'),
    ('tbl_company'),
    ('tbl_company_address'),
    ('tbl_company_contact'),
    ('tbl_fixed_asset'),
    ('tbl_hris_benefit_plan'),
    ('tbl_hris_employee'),
    ('tbl_hris_holiday'),
    ('tbl_hris_holiday_setup'),
    -- tbl_hris_payroll_item is NOT setup data despite the name - it's a payroll
    -- run's line items (FK tbl_hris_payroll_item.run_id -> tbl_hris_payroll_run.id),
    -- confirmed the hard way when this script's first run left that FK untrusted.
    ('tbl_inv_warehouse_address'),
    ('tbl_inv_warehouse_area'),
    ('tbl_inv_warehouse_name'),
    ('tbl_inv_warehouse_usetype'),
    ('tbl_item_trade_type'),
    ('tbl_position'),
    ('tbl_position_access'),
    ('tbl_setup_application'),
    ('tbl_setup_asset_category'),
    ('tbl_setup_bpi_entity'),
    ('tbl_setup_bpi_industries'),
    ('tbl_setup_calendar_category'),
    ('tbl_setup_calendar_cost_type'),
    ('tbl_setup_chart_class'),
    ('tbl_setup_chart_of_accounts'),
    ('tbl_setup_item'),
    ('tbl_setup_item_additional_specs'),
    ('tbl_setup_item_additional_specs_pump_type'),
    ('tbl_setup_item_bom'),
    ('tbl_setup_item_bom_details'),
    ('tbl_setup_item_boq'),
    ('tbl_setup_item_boq_details'),
    ('tbl_setup_item_brand'),
    ('tbl_setup_item_class'),
    ('tbl_setup_item_image'),
    ('tbl_setup_item_inventory'),
    ('tbl_setup_item_material'),
    ('tbl_setup_item_model'),
    ('tbl_setup_item_name'),
    ('tbl_setup_item_pump_count'),
    ('tbl_setup_item_pump_type'),
    ('tbl_setup_item_specs'),
    ('tbl_setup_item_specs_template'),
    ('tbl_setup_item_trade_type'),
    ('tbl_setup_item_type'),
    ('tbl_setup_item_unit_measurement'),
    ('tbl_setup_payment_terms'),
    ('tbl_setup_ship_type'),
    ('tbl_setup_social'),
    ('tbl_setup_status'),
    ('tbl_setup_tax'),
    ('tbl_setup_tax_details'),
    ('tbl_setup_users'),
    ('tbl_setup_valuation_method'),
    ('tbl_trans_sales_project_template'),        -- the Project Quotation default templates, not a customer transaction
    ('tbl_trans_sales_project_template_child'),
    ('tbl_user_permission'),
    ('tbl_vehicle'),
    ('tbl_vehicle_file');

DECLARE @sql NVARCHAR(MAX) = N'';

-- 1. Disable every FK constraint and the one known trigger DB-wide.
SELECT @sql = @sql + N'ALTER TABLE ' + QUOTENAME(s.name) + N'.' + QUOTENAME(t.name) + N' NOCHECK CONSTRAINT ALL;' + CHAR(10)
FROM sys.tables t JOIN sys.schemas s ON t.schema_id = s.schema_id;
EXEC sp_executesql @sql;

IF OBJECT_ID('dbo.tr_inv_item_stocks_ledger', 'TR') IS NOT NULL
    DISABLE TRIGGER dbo.tr_inv_item_stocks_ledger ON dbo.tbl_inv_item_stocks;

-- 2. Delete every row from every table NOT in the keep list (covers both
--    the transactional tables and every z_-prefixed audit table).
SET @sql = N'';
SELECT @sql = @sql + N'DELETE FROM ' + QUOTENAME(s.name) + N'.' + QUOTENAME(t.name) + N';' + CHAR(10)
FROM sys.tables t
JOIN sys.schemas s ON t.schema_id = s.schema_id
WHERE t.name NOT IN (SELECT name FROM #KeepTables);
EXEC sp_executesql @sql;

-- 3. Reseed IDENTITY back to 0 on every cleared table that has one, so the
--    next real insert starts a clean 1, 2, 3...
SET @sql = N'';
SELECT @sql = @sql + N'DBCC CHECKIDENT(''' + s.name + N'.' + t.name + N''', RESEED, 0);' + CHAR(10)
FROM sys.tables t
JOIN sys.schemas s ON t.schema_id = s.schema_id
JOIN sys.identity_columns ic ON ic.object_id = t.object_id
WHERE t.name NOT IN (SELECT name FROM #KeepTables);
EXEC sp_executesql @sql;

-- 4. Re-enable the trigger and every FK constraint, re-validating as we go -
--    should never fail here, since only the "many" side of any relationship
--    was ever cleared, never a kept parent row.
IF OBJECT_ID('dbo.tr_inv_item_stocks_ledger', 'TR') IS NOT NULL
    ENABLE TRIGGER dbo.tr_inv_item_stocks_ledger ON dbo.tbl_inv_item_stocks;

SET @sql = N'';
SELECT @sql = @sql + N'ALTER TABLE ' + QUOTENAME(s.name) + N'.' + QUOTENAME(t.name) + N' WITH CHECK CHECK CONSTRAINT ALL;' + CHAR(10)
FROM sys.tables t JOIN sys.schemas s ON t.schema_id = s.schema_id;
EXEC sp_executesql @sql;

DROP TABLE #KeepTables;

PRINT 'Reset complete - setup/master data preserved, all transactional and audit history cleared.';
