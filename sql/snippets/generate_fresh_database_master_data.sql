-- ============================================================================
-- generate_fresh_database_master_data.sql
-- ============================================================================
-- Writes the five numbered scripts of a "fresh database master data" batch: everything
-- a database the API has just created needs before Item Entry, BOMs and BPI can be used -
-- the setup lists, the item catalogue with its specs, the BOMs and project templates, and
-- the business partners - with the SAME ids the source uses, so nothing has to be looked
-- up or re-linked on the way in.
--
-- Run it against the SOURCE database - the one whose master data a new database should
-- start with. It only reads. One run writes one file; run it once per PART (1 to 5):
--
--   sqlcmd -S <server> -U sa -P <pw> -d <source db> -b -I -y 0 -f o:65001
--          -v PART=1 -i generate_fresh_database_master_data.sql
--          -o <YYYY-MM-DD>_NN_fresh_database_master_data\01_setup_lists_<YYYY-MM-DD>.sql
--
--   PART=1  01_setup_lists_<date>.sql
--   PART=2  02_items_<date>.sql
--   PART=3  03_item_specs_<date>.sql
--   PART=4  04_boms_and_templates_<date>.sql
--   PART=5  05_bpi_<date>.sql
--
-- Put the five in a NEW dated batch folder (sql/snippets/<date>_<NN>_fresh_database_master_data)
-- rather than over an earlier one - an earlier batch may already have been run somewhere.
--
-- The switches matter: -f o:65001 writes UTF-8 with a byte-order mark, so SSMS reads the
-- non-ASCII item models correctly; -y 0 writes every value whole - no column header,
-- no padding, no wrapping. This file uses a sqlcmd variable, so in SSMS it needs
-- Query > SQLCMD Mode.
--
-- What each generated file does is described in its own header. The loader at the
-- bottom of each file is written below with ~ standing for a single quote (so it needs
-- no escaping here); the generator swaps them back.
-- ============================================================================
SET NOCOUNT ON;
SET QUOTED_IDENTIFIER ON;

DECLARE @part int = $(PART);

IF @part NOT BETWEEN 1 AND 5
BEGIN
    RAISERROR('PART must be 1, 2, 3, 4 or 5.', 16, 1);
    RETURN;
END

DECLARE @stamp       nvarchar(10) = CONVERT(nvarchar(10), GETDATE(), 23);
DECLARE @crlf        nchar(2)     = NCHAR(13) + NCHAR(10);
DECLARE @float_style int          = CASE WHEN CAST(PARSENAME(CAST(SERVERPROPERTY('ProductVersion') AS nvarchar(40)), 4) AS int) >= 13
                                         THEN 3   -- 17 significant digits, exact round trip (2016 and later)
                                         ELSE 2 END;

-- The five files ----------------------------------------------------------------
DECLARE @files TABLE (part int PRIMARY KEY, stub nvarchar(60), title nvarchar(200), blurb nvarchar(max));
INSERT INTO @files VALUES
 (1, N'setup_lists', N'setup lists - positions, users, company, chart of accounts, tax, and every item and BPI lookup',
     N'-- The lists everything else points at: positions and users (with their existing passwords),' + @crlf
   + N'-- Company Setup, the chart of accounts and tax codes, payment terms, and every list Item' + @crlf
   + N'-- Entry and BPI pick from (item name, class, brand, unit, pump type and count, material...).' + @crlf
   + N'-- The chart of accounts keeps its ids on purpose: posting code refers to some accounts by id.'),
 (2, N'items', N'Item Entry - the item catalogue',
     N'-- Every item with its trade types, inventory settings, models and image records.' + @crlf
   + N'-- Image FILES are not in the database - copy the API''s files folder across separately.'),
 (3, N'item_specs', N'item specs - long descriptions, pump type and count, spec templates',
     N'-- The Specifications and Additional Specs tabs of Item Entry for every item.'),
 (4, N'boms_and_templates', N'BOMs, BOQs and project quotation templates',
     N'-- Every production BOM with its component lines and line remarks, any BOQs, and the' + @crlf
   + N'-- project quotation templates from Template Setup.'),
 (5, N'bpi', N'BPI - business partners',
     N'-- Every customer and supplier: branches, addresses, contacts, industries, entity, finance' + @crlf
   + N'-- terms, supplier items, accreditation records and history. Accreditation FILES are not in' + @crlf
   + N'-- the database - copy the API''s files folder across separately.');

-- Which table goes in which file, in load order (parents before children) ----------
-- placeholder: the rows the API seeds into a brand-new database (seed_defaults.go and
-- seed_position_access.go). The loader deletes exactly those and refuses to touch a table
-- holding anything else. Keep these in step with the seeds if they change.
DECLARE @ph_position nvarchar(max) = N'code IN (N''DISPATCH'', N''CASHIER'') AND name IN (N''Dispatcher'', N''AR/AP Cashier'')';
DECLARE @ph_template nvarchar(max) = N'template_name = N''DEFAULT STARTER TEMPLATE (PLACEHOLDER)''';

IF OBJECT_ID('tempdb..#pick') IS NOT NULL DROP TABLE #pick;
CREATE TABLE #pick (part int NOT NULL, ord int NOT NULL, tbl sysname NOT NULL PRIMARY KEY,
                    placeholder nvarchar(max) NULL, delete_only bit NOT NULL, rows_ int NULL);

INSERT INTO #pick (part, ord, tbl, placeholder, delete_only) VALUES
 (1,  10, N'tbl_position',                    @ph_position, 0),
 (1,  20, N'tbl_position_access',             N'position_id IN (SELECT id FROM dbo.tbl_position WHERE ' + @ph_position + N')', 1),
 (1,  30, N'tbl_setup_users',                 NULL, 0),
 (1,  40, N'tbl_company',                     NULL, 0),
 (1,  50, N'tbl_setup_chart_class',           NULL, 0),
 (1,  60, N'tbl_setup_chart_of_accounts',     NULL, 0),
 (1,  70, N'tbl_setup_tax',                   NULL, 0),
 (1,  80, N'tbl_setup_tax_details',           NULL, 0),
 (1,  90, N'tbl_setup_payment_terms',         NULL, 0),
 (1, 100, N'tbl_setup_item_name',             N'code = N''DEFAULT'' AND name = N''Default Item Name (Placeholder)''', 0),
 (1, 110, N'tbl_setup_item_class',            N'code = N''DEFAULT'' AND name = N''Default (Placeholder)''', 0),
 (1, 120, N'tbl_setup_item_brand',            N'code = N''DEFAULT'' AND name = N''Default (Placeholder)''', 0),
 (1, 130, N'tbl_setup_item_unit_measurement', N'code = N''UNIT'' AND name = N''Unit (Placeholder)''', 0),
 (1, 140, N'tbl_setup_item_type',             NULL, 0),
 (1, 150, N'tbl_setup_item_material',         NULL, 0),
 (1, 160, N'tbl_setup_item_pump_type',        NULL, 0),
 (1, 170, N'tbl_setup_item_pump_count',       NULL, 0),
 (1, 180, N'tbl_setup_valuation_method',      NULL, 0),
 (1, 190, N'tbl_inv_warehouse_usetype',       NULL, 0),
 (2,  10, N'tbl_setup_item',                  N'item_code = N''DEFAULT-PLACEHOLDER'' AND item_model = N''DEFAULT-PLACEHOLDER''', 0),
 (2, 900, N'tbl_item_trade_type',             NULL, 0),
 (3,  10, N'tbl_setup_item_additional_specs', NULL, 0),
 (3,  20, N'tbl_setup_item_additional_specs_pump_type', NULL, 0),
 (3,  30, N'tbl_setup_item_specs',            NULL, 0),
 (3,  40, N'tbl_setup_item_specs_template',   NULL, 0),
 (4,  10, N'tbl_setup_item_bom',              NULL, 0),
 (4,  20, N'tbl_setup_item_bom_details',      NULL, 0),
 (4,  30, N'tbl_setup_item_boq',              NULL, 0),
 (4,  40, N'tbl_setup_item_boq_details',      NULL, 0),
 (4,  50, N'tbl_boq_notes',                   NULL, 0),
 (4,  60, N'tbl_trans_sales_project_template', @ph_template, 0),
 (4,  70, N'tbl_trans_sales_project_template_child',
          N'parent_id IN (SELECT template_id FROM dbo.tbl_trans_sales_project_template WHERE ' + @ph_template + N')', 0),
 (5,  10, N'tbl_bpi',                         NULL, 0),
 (5,  20, N'tbl_bpi_general',                 NULL, 0);

-- Everything else of the same families lands in a file too, so a setup list added later is
-- not silently left behind. Not carried at all: tbl_setup_calendar_cost_type, which the API
-- seeds itself; transactions; the z_*_at audit copies; and zz_* backups.
INSERT INTO #pick (part, ord, tbl, placeholder, delete_only)
SELECT CASE WHEN t.name LIKE N'tbl[_]bpi%'                THEN 5
            WHEN t.name LIKE N'tbl[_]setup[_]item[_]bo[mq]%' THEN 4
            WHEN t.name LIKE N'tbl[_]setup[_]item[_]%'    THEN 2
            ELSE 1 END,
       1000 + ROW_NUMBER() OVER (ORDER BY t.name),
       t.name, NULL, 0
  FROM sys.tables t
 WHERE (t.name LIKE N'tbl[_]setup[_]%' OR t.name LIKE N'tbl[_]bpi%')
   AND t.name NOT IN (N'tbl_setup_calendar_cost_type')
   AND NOT EXISTS (SELECT 1 FROM #pick p WHERE p.tbl = t.name);

DELETE FROM #pick WHERE part <> @part OR OBJECT_ID(N'dbo.' + QUOTENAME(tbl), N'U') IS NULL;

DECLARE @t sysname, @sql nvarchar(max), @n int;
DECLARE cnt_cur CURSOR LOCAL FAST_FORWARD FOR SELECT tbl FROM #pick;
OPEN cnt_cur;
FETCH NEXT FROM cnt_cur INTO @t;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N'SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@t) + N';';
    EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
    UPDATE #pick SET rows_ = @n WHERE tbl = @t;
    FETCH NEXT FROM cnt_cur INTO @t;
END
CLOSE cnt_cur;
DEALLOCATE cnt_cur;

-- An empty source table leaves the new database's table as the API made it.
DELETE FROM #pick WHERE rows_ = 0 AND delete_only = 0;

-- Only the column types this generator knows how to write out. Anything else stops it
-- rather than producing a literal that silently loses precision.
IF EXISTS (SELECT 1 FROM #pick p JOIN sys.columns c ON c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(p.tbl))
                                 JOIN sys.types ty ON ty.user_type_id = c.user_type_id
            WHERE p.delete_only = 0
              AND ty.name NOT IN (N'bigint', N'int', N'smallint', N'tinyint', N'bit', N'float', N'real', N'decimal', N'numeric',
                                  N'nvarchar', N'nchar', N'varchar', N'char', N'date', N'datetime', N'datetime2', N'datetimeoffset'))
BEGIN
    SELECT p.tbl, c.name, ty.name FROM #pick p JOIN sys.columns c ON c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(p.tbl))
      JOIN sys.types ty ON ty.user_type_id = c.user_type_id
     WHERE ty.name NOT IN (N'bigint', N'int', N'smallint', N'tinyint', N'bit', N'float', N'real', N'decimal', N'numeric',
                           N'nvarchar', N'nchar', N'varchar', N'char', N'date', N'datetime', N'datetime2', N'datetimeoffset');
    RAISERROR('A column type above is not supported by this generator - extend it before generating.', 16, 1);
    RETURN;
END

-- References the loader re-checks after loading ------------------------------------
-- child.col -> parent.pcol, 0 and NULL meaning "none". Counted here in the source and
-- written into the file; the loader fails only if MORE rows dangle after the load.
DECLARE @refs TABLE (child sysname, col sysname, parent sysname, pcol sysname);
INSERT INTO @refs VALUES
 (N'tbl_setup_users',                  N'position_id',                 N'tbl_position',                    N'id'),
 (N'tbl_setup_chart_of_accounts',      N'class_id',                    N'tbl_setup_chart_class',           N'id'),
 (N'tbl_setup_tax',                    N'coa_sales_id',                N'tbl_setup_chart_of_accounts',     N'id'),
 (N'tbl_setup_tax',                    N'coa_purchase_id',             N'tbl_setup_chart_of_accounts',     N'id'),
 (N'tbl_setup_tax_details',            N'tax_code_id',                 N'tbl_setup_tax',                   N'id'),
 (N'tbl_setup_item',                   N'item_name_id',                N'tbl_setup_item_name',             N'id'),
 (N'tbl_setup_item',                   N'item_class_id',               N'tbl_setup_item_class',            N'id'),
 (N'tbl_setup_item',                   N'item_brand_id',               N'tbl_setup_item_brand',            N'id'),
 (N'tbl_setup_item',                   N'unit_of_measure_id',          N'tbl_setup_item_unit_measurement', N'id'),
 (N'tbl_setup_item_trade_type',        N'item_id',                     N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_trade_type',        N'trade_type_id',               N'tbl_setup_item_type',             N'id'),
 (N'tbl_item_trade_type',              N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_image',             N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_inventory',         N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_inventory',         N'valuation_method_id',         N'tbl_setup_valuation_method',      N'id'),
 (N'tbl_setup_item_inventory',         N'warehouse_id',                N'tbl_inv_warehouse_name',          N'id'),
 (N'tbl_setup_item_model',             N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_additional_specs',  N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_additional_specs',  N'material_id',                 N'tbl_setup_item_material',         N'id'),
 (N'tbl_setup_item_additional_specs',  N'pump_count_compatability_id', N'tbl_setup_item_pump_count',       N'id'),
 (N'tbl_setup_item_additional_specs',  N'volume_unit_of_measure_id',   N'tbl_setup_item_unit_measurement', N'id'),
 (N'tbl_setup_item_additional_specs',  N'weight_unit_of_measure_id',   N'tbl_setup_item_unit_measurement', N'id'),
 (N'tbl_setup_item_additional_specs_pump_type', N'additional_specs_id',        N'tbl_setup_item_additional_specs', N'id'),
 (N'tbl_setup_item_additional_specs_pump_type', N'pump_type_compatability_id', N'tbl_setup_item_pump_type',        N'id'),
 (N'tbl_setup_item_specs',             N'based_id',                    N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_specs',             N'impeller_id',                 N'tbl_setup_item_material',         N'id'),
 (N'tbl_setup_item_specs_template',    N'based_id',                    N'tbl_setup_item_specs',            N'id'),
 (N'tbl_setup_item_bom',               N'item_id',                     N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_bom_details',       N'item_bom_id',                 N'tbl_setup_item_bom',              N'id'),
 (N'tbl_setup_item_bom_details',       N'item_id',                     N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_boq',               N'item_id',                     N'tbl_setup_item',                  N'id'),
 (N'tbl_setup_item_boq_details',       N'qq_id',                       N'tbl_setup_item_boq',              N'id'),
 (N'tbl_trans_sales_project_template_child', N'parent_id',             N'tbl_trans_sales_project_template', N'template_id'),
 (N'tbl_trans_sales_project_template_child', N'item_id',               N'tbl_setup_item',                  N'id'),
 (N'tbl_bpi_general',                  N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_general',                  N'social',                      N'tbl_setup_social',                N'id'),
 (N'tbl_bpi_address',                  N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_address',                  N'branch_id',                   N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_contacts',                 N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_contacts',                 N'branch_id',                   N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_contacts',                 N'position',                    N'tbl_position',                    N'id'),
 (N'tbl_bpi_accreditation',            N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_accreditation',            N'branch_id',                   N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_history',                  N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_finance',                  N'finance_based_id',            N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_finance',                  N'finance_branch_id',           N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_finance',                  N'finance_account_id',          N'tbl_setup_chart_of_accounts',     N'id'),
 (N'tbl_bpi_finance',                  N'finance_payment_terms_id',    N'tbl_setup_payment_terms',         N'id'),
 (N'tbl_bpi_items',                    N'based_id',                    N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_items',                    N'branch_id',                   N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_items',                    N'item_id',                     N'tbl_setup_item',                  N'id'),
 (N'tbl_bpi_items',                    N'payment_terms_id',            N'tbl_setup_payment_terms',         N'id'),
 (N'tbl_bpi_items',                    N'item_account_id',             N'tbl_setup_chart_of_accounts',     N'id'),
 (N'tbl_bpi_industries',               N'bpi_id',                      N'tbl_bpi',                         N'id'),
 (N'tbl_bpi_industries',               N'industry_id',                 N'tbl_setup_bpi_industries',        N'id'),
 (N'tbl_bpi_branch_industries',        N'bpi_general_id',              N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_branch_industries',        N'industry_id',                 N'tbl_setup_bpi_industries',        N'id'),
 (N'tbl_bpi_entity',                   N'bpi_general_id',              N'tbl_bpi_general',                 N'id'),
 (N'tbl_bpi_entity',                   N'entity_id',                   N'tbl_setup_bpi_entity',            N'id');

IF OBJECT_ID('tempdb..#refcount') IS NOT NULL DROP TABLE #refcount;
CREATE TABLE #refcount (child sysname, col sysname, parent sysname, pcol sysname, dangling int);

DECLARE @child sysname, @col sysname, @parent sysname, @pcol sysname;
DECLARE ref_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT r.child, r.col, r.parent, r.pcol FROM @refs r WHERE EXISTS (SELECT 1 FROM #pick p WHERE p.tbl = r.child);
OPEN ref_cur;
FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
WHILE @@FETCH_STATUS = 0
BEGIN
    IF COL_LENGTH(N'dbo.' + QUOTENAME(@child), @col) IS NOT NULL
    BEGIN
        IF COL_LENGTH(N'dbo.' + QUOTENAME(@parent), @pcol) IS NOT NULL
        BEGIN
            SET @sql = N'SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@child) + N' c WHERE c.' + QUOTENAME(@col)
                     + N' <> 0 AND NOT EXISTS (SELECT 1 FROM dbo.' + QUOTENAME(@parent) + N' p WHERE p.' + QUOTENAME(@pcol)
                     + N' = c.' + QUOTENAME(@col) + N');';
            EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
        END
        ELSE
        BEGIN
            -- The parent table is not in the source at all: every non-zero value dangles.
            SET @sql = N'SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@child) + N' WHERE ' + QUOTENAME(@col) + N' <> 0;';
            EXEC sp_executesql @sql, N'@n int OUTPUT', @n = @n OUTPUT;
        END
        INSERT INTO #refcount VALUES (@child, @col, @parent, @pcol, @n);
    END
    FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
END
CLOSE ref_cur;
DEALLOCATE ref_cur;

-- The file ------------------------------------------------------------------------
IF OBJECT_ID('tempdb..#out') IS NOT NULL DROP TABLE #out;
CREATE TABLE #out (n int IDENTITY(1,1) PRIMARY KEY, txt nvarchar(max) NOT NULL);

DECLARE @stub nvarchar(60), @title nvarchar(200), @blurb nvarchar(max), @next_stub nvarchar(60);
SELECT @stub = stub, @title = title, @blurb = blurb FROM @files WHERE part = @part;
SELECT @next_stub = stub FROM @files WHERE part = @part + 1;

DECLARE @pp nchar(2) = RIGHT(N'0' + CAST(@part AS nvarchar(2)), 2);

INSERT INTO #out (txt) VALUES
 (N'-- STEP ' + @pp + N' of 05 - ' + @title),
 (N'-- Run the numbered scripts in this folder in order, each on its own: 01, 02 ... 05.'),
 (N'-- Run after:  ' + CASE WHEN @part = 1 THEN N'starting the API once against the NEW database, so it builds the tables'
                          ELSE (SELECT RIGHT(N'0' + CAST(part AS nvarchar(2)), 2) + N'_' + stub + N'_' + @stamp + N'.sql' FROM @files WHERE part = @part - 1) END),
 (N'-- Run before: ' + CASE WHEN @part = 5 THEN N'restarting the API (see AFTER THE LAST FILE below)'
                          ELSE RIGHT(N'0' + CAST(@part + 1 AS nvarchar(2)), 2) + N'_' + @next_stub + N'_' + @stamp + N'.sql' END),
 (N'-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".'),
 (N''),
 (N'-- Master data for a NEW database: ' + @title),
 (N'-- ============================================================================'),
 (@blurb),
 (N'--'),
 (N'-- Generated ' + CONVERT(nvarchar(16), GETDATE(), 120) + N' from ' + @@SERVERNAME + N' / ' + DB_NAME()
   + N' by sql/snippets/generate_fresh_database_master_data.sql.'),
 (N'-- Regenerate rather than edit: the data below is a snapshot of that database on that day.'),
 (N'--'),
 (N'-- WHAT THE BATCH IS FOR'),
 (N'--   A database the API has just created has its tables but no master data. The five files'),
 (N'--   of this batch give it the source''s master data with the SAME ids the source uses, so every'),
 (N'--   reference between them (item -> brand, BOM line -> item, BPI -> payment terms...) still lines'),
 (N'--   up and scripts written against the source work here too. NOT carried: transactions (quotes,'),
 (N'--   orders, stock, accounting documents), the z_*_at audit history, and what the API seeds for'),
 (N'--   itself (access modules, position access, calendar cost types).'),
 (N'--'),
 (N'-- WHAT IT REFUSES TO DO'),
 (N'--   It loads into EMPTY tables only. The rows the API seeds into a new database - the DEFAULT'),
 (N'--   item name, class and brand, the UNIT placeholder, the DEFAULT-PLACEHOLDER item, the starter'),
 (N'--   template, and the Dispatcher and AR/AP Cashier positions with their access - are removed and'),
 (N'--   replaced. Anything else already in a table stops the script before it writes a thing, so it'),
 (N'--   cannot damage a working database, and running a committed file again changes nothing.'),
 (N'--'),
 (N'-- WHEN THE SCHEMA HAS MOVED ON'),
 (N'--   Columns are matched by NAME. A new nullable column in the database is simply left empty.'),
 (N'--   It stops, listing what and why, for: a table the database does not have (set'),
 (N'--   @skip_missing_tables = 1 to load without it); a column holding data here that the database'),
 (N'--   does not have (@allow_dropped_columns = 1 to load without it); a NOT NULL column with no'),
 (N'--   default that this file cannot fill; text longer than its column; a fraction headed for a'),
 (N'--   whole-number column. After loading it re-checks every reference and rolls everything back'),
 (N'--   if more rows point at nothing than did in the source.'),
 (N'--'),
 (N'-- AFTER THE LAST FILE'),
 (N'--   Restart the API. On start it re-creates the Dispatcher and AR/AP Cashier positions,'),
 (N'--   gives every position its starting access, and the item and BPI lists are read fresh.'),
 (N'--'),
 (N'-- Tables in this file, in load order (rows):');

INSERT INTO #out (txt)
SELECT N'--   ' + tbl + CASE WHEN delete_only = 1 THEN N'  (API placeholder rows removed only)'
                        ELSE N' (' + CAST(rows_ AS nvarchar(20)) + N')' END
  FROM #pick ORDER BY ord;

INSERT INTO #out (txt) VALUES
 (N'--'),
 (N'-- HOW TO RUN: open it in SSMS on the NEW database and execute, or'),
 (N'--   sqlcmd -S <server> -d <database> -b -i <this file>'),
 (N'-- With @commit = 0 everything is checked, loaded and counted, then rolled back.'),
 (N''),
 (N'SET NOCOUNT ON;'),
 (N'SET XACT_ABORT ON;'),
 (N''),
 (N'DECLARE @commit bit = 0;                 -- 1 = keep the changes, 0 = trial run'),
 (N'DECLARE @skip_missing_tables bit = 0;    -- see WHEN THE SCHEMA HAS MOVED ON'),
 (N'DECLARE @allow_dropped_columns bit = 0;  -- see WHEN THE SCHEMA HAS MOVED ON'),
 (N''),
 (N'IF OBJECT_ID(''tempdb..#plan'') IS NOT NULL DROP TABLE #plan;'),
 (N'IF OBJECT_ID(''tempdb..#refs'') IS NOT NULL DROP TABLE #refs;');

INSERT INTO #out (txt)
SELECT N'IF OBJECT_ID(''tempdb..#m_' + tbl + N''') IS NOT NULL DROP TABLE #m_' + tbl + N';'
  FROM #pick WHERE delete_only = 0 ORDER BY ord;

INSERT INTO #out (txt) VALUES
 (N''),
 (N'-- What goes where. placeholder = the API''s own rows, the only ones this may delete.'),
 (N'CREATE TABLE #plan (ord int NOT NULL PRIMARY KEY, target sysname NOT NULL, stage sysname NULL,'),
 (N'                    rows_in_file int NOT NULL, placeholder nvarchar(max) NULL);'),
 (N'INSERT INTO #plan (ord, target, stage, rows_in_file, placeholder) VALUES');

INSERT INTO #out (txt)
SELECT N'    (' + CAST(ord AS nvarchar(10)) + N', N''' + tbl + N''', '
     + CASE WHEN delete_only = 1 THEN N'NULL' ELSE N'N''#m_' + tbl + N'''' END + N', '
     + CAST(CASE WHEN delete_only = 1 THEN 0 ELSE rows_ END AS nvarchar(20)) + N', '
     + ISNULL(N'N''' + REPLACE(placeholder, N'''', N'''''') + N'''', N'NULL') + N')'
     + CASE WHEN ROW_NUMBER() OVER (ORDER BY ord DESC) = 1 THEN N';' ELSE N',' END
  FROM #pick ORDER BY ord;

INSERT INTO #out (txt) VALUES
 (N''),
 (N'-- References re-checked after loading, with how many already dangled in the source.'),
 (N'CREATE TABLE #refs (child sysname NOT NULL, col sysname NOT NULL, parent sysname NOT NULL, pcol sysname NOT NULL,'),
 (N'                    dangling_in_source int NOT NULL, dangling_here int NULL);');

IF EXISTS (SELECT 1 FROM #refcount)
BEGIN
    INSERT INTO #out (txt) VALUES (N'INSERT INTO #refs (child, col, parent, pcol, dangling_in_source) VALUES');
    INSERT INTO #out (txt)
    SELECT N'    (N''' + child + N''', N''' + col + N''', N''' + parent + N''', N''' + pcol + N''', '
         + CAST(dangling AS nvarchar(20)) + N')'
         + CASE WHEN ROW_NUMBER() OVER (ORDER BY child DESC, col DESC) = 1 THEN N';' ELSE N',' END
      FROM #refcount ORDER BY child, col;
END

-- The data: one staging table per source table, filled 500 rows per INSERT -----------
DECLARE @ord int, @rows int, @defs nvarchar(max), @cols nvarchar(max), @vals nvarchar(max), @orderby nvarchar(300), @ins nvarchar(max);
DECLARE @dol nvarchar(2) = N'$' + N'(';   -- built in two parts: written whole, sqlcmd would take it for a variable

DECLARE tbl_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, tbl, rows_ FROM #pick WHERE delete_only = 0 ORDER BY ord;
OPEN tbl_cur;
FETCH NEXT FROM tbl_cur INTO @ord, @t, @rows;
WHILE @@FETCH_STATUS = 0
BEGIN
    -- Staging columns take the source's own types; every one nullable.
    SET @defs = STUFF((
        SELECT N', ' + QUOTENAME(c.name) + N' '
             + CASE ty.name
                    WHEN N'nvarchar'  THEN N'nvarchar(' + CASE WHEN c.max_length = -1 THEN N'max' ELSE CAST(c.max_length / 2 AS nvarchar(10)) END + N')'
                    WHEN N'nchar'     THEN N'nchar(' + CAST(c.max_length / 2 AS nvarchar(10)) + N')'
                    WHEN N'varchar'   THEN N'varchar(' + CASE WHEN c.max_length = -1 THEN N'max' ELSE CAST(c.max_length AS nvarchar(10)) END + N')'
                    WHEN N'char'      THEN N'char(' + CAST(c.max_length AS nvarchar(10)) + N')'
                    WHEN N'decimal'   THEN N'decimal(' + CAST(c.precision AS nvarchar(5)) + N', ' + CAST(c.scale AS nvarchar(5)) + N')'
                    WHEN N'numeric'   THEN N'decimal(' + CAST(c.precision AS nvarchar(5)) + N', ' + CAST(c.scale AS nvarchar(5)) + N')'
                    WHEN N'datetime2' THEN N'datetime2(' + CAST(c.scale AS nvarchar(5)) + N')'
                    WHEN N'datetimeoffset' THEN N'datetimeoffset(' + CAST(c.scale AS nvarchar(5)) + N')'
                    ELSE ty.name END
             + N' NULL'
          FROM sys.columns c JOIN sys.types ty ON ty.user_type_id = c.user_type_id
         WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@t))
         ORDER BY c.column_id
        FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, N'');

    SET @cols = STUFF((
        SELECT N', ' + QUOTENAME(c.name)
          FROM sys.columns c WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@t))
         ORDER BY c.column_id
        FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, N'');

    -- One expression per column rendering a T-SQL literal. The literals it needs arrive as
    -- parameters (@null, @nq, @q ...) so the expression itself contains no quotes at all.
    --   text: N'...' with quotes doubled; CR and LF written as NCHAR(13)/NCHAR(10) so every row
    --         stays on one line; a dollar sign followed by an opening bracket is split in two, so
    --         sqlcmd never reads it as one of its own variables.
    SET @vals = STUFF((
        SELECT N' + @sep + '
             + CASE
                 WHEN ty.name IN (N'nvarchar', N'nchar', N'varchar', N'char')
                   THEN N'CASE WHEN ' + QUOTENAME(c.name) + N' IS NULL THEN @null ELSE @nq + REPLACE(REPLACE(REPLACE(REPLACE(CAST('
                      + QUOTENAME(c.name) + N' AS nvarchar(max)), @q, @q2), NCHAR(13), @cr), NCHAR(10), @lf), @dol, @dolrep) + @q END'
                 -- CASE, not ISNULL: ISNULL takes its type from its first argument, so an
                 -- nvarchar(1) there cuts the NULL it falls back to down to "N".
                 WHEN ty.name = N'bit'
                   THEN N'CASE WHEN ' + QUOTENAME(c.name) + N' IS NULL THEN @null ELSE CAST(CAST(' + QUOTENAME(c.name) + N' AS int) AS nvarchar(1)) END'
                 -- A whole number is written as one (exact, and a fraction of the size);
                 -- anything else with every significant digit.
                 WHEN ty.name IN (N'float', N'real')
                   THEN N'CASE WHEN ' + QUOTENAME(c.name) + N' IS NULL THEN @null WHEN ' + QUOTENAME(c.name) + N' = FLOOR(' + QUOTENAME(c.name)
                      + N') AND ABS(' + QUOTENAME(c.name) + N') < 1e15 THEN CAST(CAST(' + QUOTENAME(c.name) + N' AS bigint) AS nvarchar(30)) ELSE CONVERT(nvarchar(40), CAST('
                      + QUOTENAME(c.name) + N' AS float), ' + CAST(@float_style AS nvarchar(2)) + N') END'
                 WHEN ty.name IN (N'date', N'datetime', N'datetime2', N'datetimeoffset')
                   THEN N'CASE WHEN ' + QUOTENAME(c.name) + N' IS NULL THEN @null ELSE @nq + CONVERT(nvarchar(40), ' + QUOTENAME(c.name) + N', 126) + @q END'
                 ELSE N'CASE WHEN ' + QUOTENAME(c.name) + N' IS NULL THEN @null ELSE CAST(' + QUOTENAME(c.name) + N' AS nvarchar(60)) END'
               END
          FROM sys.columns c JOIN sys.types ty ON ty.user_type_id = c.user_type_id
         WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@t))
         ORDER BY c.column_id
        FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 10, N'');

    SET @orderby = ISNULL(
        (SELECT QUOTENAME(c.name) FROM sys.columns c WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@t)) AND c.is_identity = 1),
        (SELECT QUOTENAME(c.name) FROM sys.columns c WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@t)) AND c.column_id = 1));

    SET @ins = N'INSERT INTO #m_' + @t + N' (' + @cols + N') VALUES' + @crlf;

    INSERT INTO #out (txt) VALUES
     (N''),
     (N'-- ' + @t + N' (' + CAST(@rows AS nvarchar(20)) + N' rows) ' + REPLICATE(N'-', CASE WHEN 60 - LEN(@t) > 3 THEN 60 - LEN(@t) ELSE 3 END)),
     (N'CREATE TABLE #m_' + @t + N' (' + @defs + N');');

    SET @sql = N'INSERT INTO #out (txt) SELECT CASE WHEN (x.[__rn] - 1) % 500 = 0 THEN @ins ELSE @empty END + @open + '
             + @vals
             + N' + @close + CASE WHEN x.[__rn] % 500 = 0 OR x.[__rn] = x.[__cnt] THEN @semi ELSE @comma END'
             + N' FROM (SELECT *, ROW_NUMBER() OVER (ORDER BY ' + @orderby + N') AS [__rn], COUNT(*) OVER () AS [__cnt]'
             + N' FROM dbo.' + QUOTENAME(@t) + N') AS x ORDER BY x.[__rn];';

    EXEC sp_executesql @sql,
        N'@ins nvarchar(max), @empty nvarchar(1), @open nvarchar(10), @close nvarchar(1), @semi nvarchar(1), @comma nvarchar(1),
          @sep nvarchar(2), @null nvarchar(4), @nq nvarchar(2), @q nvarchar(1), @q2 nvarchar(2), @cr nvarchar(20), @lf nvarchar(20),
          @dol nvarchar(2), @dolrep nvarchar(20)',
        @ins = @ins, @empty = N'', @open = N'    (', @close = N')', @semi = N';', @comma = N',',
        @sep = N', ', @null = N'NULL', @nq = N'N''', @q = N'''', @q2 = N'''''',
        @cr = N''' + NCHAR(13) + N''', @lf = N''' + NCHAR(10) + N''',
        @dol = @dol, @dolrep = N'$'' + N''(';

    FETCH NEXT FROM tbl_cur INTO @ord, @t, @rows;
END
CLOSE tbl_cur;
DEALLOCATE tbl_cur;

-- The loader (identical in every file) -------------------------------------------
DECLARE @engine nvarchar(max) = N'-- ============================================================================
-- THE LOADER - the same in every file of this batch; only the data above differs.
-- It checks everything first, and writes nothing unless every check passes.
-- ============================================================================
IF OBJECT_ID(~tempdb..#problems~) IS NOT NULL DROP TABLE #problems;
IF OBJECT_ID(~tempdb..#notes~)    IS NOT NULL DROP TABLE #notes;
IF OBJECT_ID(~tempdb..#cols~)     IS NOT NULL DROP TABLE #cols;
IF OBJECT_ID(~tempdb..#loaded~)   IS NOT NULL DROP TABLE #loaded;
CREATE TABLE #problems (target sysname NOT NULL, problem nvarchar(2000) NOT NULL);
CREATE TABLE #notes    (target sysname NOT NULL, note nvarchar(2000) NOT NULL);
CREATE TABLE #cols     (ord int NOT NULL, col sysname NOT NULL, colid int NOT NULL, load_it bit NOT NULL);
CREATE TABLE #loaded   (ord int NOT NULL PRIMARY KEY, target sysname NOT NULL, rows_in_file int NOT NULL,
                        placeholders_removed int NOT NULL, rows_loaded int NOT NULL);

DECLARE @ord int, @target sysname, @stage sysname, @rows_in_file int, @placeholder nvarchar(max),
        @tobj int, @sobj int, @sql nvarchar(max), @n int, @m int, @col sysname,
        @ttype sysname, @tlen int, @stype sysname, @collist nvarchar(max), @identity bit,
        @child sysname, @parent sysname, @pcol sysname;

-- 1. Check ----------------------------------------------------------------------
DECLARE plan_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, stage, rows_in_file, placeholder FROM #plan ORDER BY ord;
OPEN plan_cur;
FETCH NEXT FROM plan_cur INTO @ord, @target, @stage, @rows_in_file, @placeholder;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @tobj = OBJECT_ID(N~dbo.~ + QUOTENAME(@target), N~U~);
    SET @sobj = OBJECT_ID(N~tempdb..~ + @stage);

    -- Every row of this file has to have reached its staging table. A run that stopped
    -- part-way (SSMS carries on after an error) would otherwise load a short table.
    IF @stage IS NOT NULL
    BEGIN
        SET @n = NULL;
        IF @sobj IS NOT NULL
        BEGIN
            SET @sql = N~SELECT @n = COUNT(*) FROM ~ + @stage + N~;~;
            EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
        END
        IF ISNULL(@n, -1) <> @rows_in_file
            INSERT INTO #problems VALUES (@target, N~only ~ + ISNULL(CAST(@n AS nvarchar(20)), N~0~) + N~ of the ~
                + CAST(@rows_in_file AS nvarchar(20)) + N~ rows in this file were read - run the whole file again, from the top~);
    END

    IF @tobj IS NULL
    BEGIN
        IF @stage IS NULL
            INSERT INTO #notes VALUES (@target, N~not in this database - nothing to clear~);
        ELSE IF @skip_missing_tables = 1
            INSERT INTO #notes VALUES (@target, N~NOT LOADED - the table is not in this database (~
                + CAST(@rows_in_file AS nvarchar(20)) + N~ rows left out because @skip_missing_tables = 1)~);
        ELSE
            INSERT INTO #problems VALUES (@target, N~the table is not in this database. Start the API against it once so it ~
                + N~builds its tables. If the table was retired on purpose, set @skip_missing_tables = 1 to load without it.~);
    END
    ELSE
    BEGIN
        -- Anything in the table except the API~s own placeholder rows means this is not a new database.
        SET @sql = N~SELECT @n = COUNT(*) FROM dbo.~ + QUOTENAME(@target) + N~; SELECT @m = COUNT(*) FROM dbo.~
                 + QUOTENAME(@target) + N~ WHERE ~ + ISNULL(@placeholder, N~1 = 0~) + N~;~;
        EXEC sp_executesql @sql, N~@n int OUTPUT, @m int OUTPUT~, @n = @n OUTPUT, @m = @m OUTPUT;
        IF @n > @m
            INSERT INTO #problems VALUES (@target, CAST(@n - @m AS nvarchar(20)) + N~ row(s) already in the table. This loads a NEW ~
                + N~database only - it never adds to or overwrites data that is there. If this file was already committed ~
                + N~here, there is nothing left to do.~);

        IF @sobj IS NOT NULL
        BEGIN
            INSERT INTO #cols (ord, col, colid, load_it)
            SELECT @ord, s.name, s.column_id,
                   CASE WHEN EXISTS (SELECT 1 FROM sys.columns t
                                      WHERE t.object_id = @tobj AND t.is_computed = 0
                                        AND t.name = s.name COLLATE DATABASE_DEFAULT) THEN 1 ELSE 0 END
              FROM tempdb.sys.columns s
             WHERE s.object_id = @sobj;

            -- A column this file carries that the database no longer has.
            DECLARE col_cur CURSOR LOCAL FAST_FORWARD FOR
                SELECT col FROM #cols WHERE ord = @ord AND load_it = 0 ORDER BY colid;
            OPEN col_cur;
            FETCH NEXT FROM col_cur INTO @col;
            WHILE @@FETCH_STATUS = 0
            BEGIN
                SET @sql = N~SELECT @n = COUNT(*) FROM ~ + @stage + N~ WHERE ~ + QUOTENAME(@col) + N~ IS NOT NULL;~;
                EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
                IF @n > 0 AND @allow_dropped_columns = 0
                    INSERT INTO #problems VALUES (@target, N~column ~ + @col + N~ holds ~ + CAST(@n AS nvarchar(20))
                        + N~ value(s) but this database has no such column. Set @allow_dropped_columns = 1 to load without it.~);
                ELSE
                    INSERT INTO #notes VALUES (@target, N~column ~ + @col + N~ left out - this database has no such column (~
                        + CAST(@n AS nvarchar(20)) + N~ value(s) not loaded)~);
                FETCH NEXT FROM col_cur INTO @col;
            END
            CLOSE col_cur;
            DEALLOCATE col_cur;

            -- A column the database insists on that this file does not carry.
            INSERT INTO #problems
            SELECT @target, N~column ~ + t.name + N~ must be filled (NOT NULL, no default) but this file has no such column~
              FROM sys.columns t
             WHERE t.object_id = @tobj AND t.is_nullable = 0 AND t.is_identity = 0 AND t.is_computed = 0
               AND t.default_object_id = 0
               AND NOT EXISTS (SELECT 1 FROM #cols c
                                WHERE c.ord = @ord AND c.load_it = 1 AND c.col = t.name COLLATE DATABASE_DEFAULT);

            -- Values that would not survive the trip: text longer than the column, or a
            -- fraction headed for a whole-number column. Both would be cut, not refused.
            DECLARE fit_cur CURSOR LOCAL FAST_FORWARD FOR
                SELECT c.col, tt.name, t.max_length, st.name
                  FROM #cols c
                  JOIN sys.columns t  ON t.object_id = @tobj AND t.name = c.col COLLATE DATABASE_DEFAULT
                  JOIN sys.types   tt ON tt.user_type_id = t.user_type_id
                  JOIN tempdb.sys.columns s ON s.object_id = @sobj AND s.column_id = c.colid
                  JOIN tempdb.sys.types  st ON st.user_type_id = s.user_type_id
                 WHERE c.ord = @ord AND c.load_it = 1;
            OPEN fit_cur;
            FETCH NEXT FROM fit_cur INTO @col, @ttype, @tlen, @stype;
            WHILE @@FETCH_STATUS = 0
            BEGIN
                SET @sql = NULL;
                IF @ttype IN (N~nvarchar~, N~nchar~) AND @tlen <> -1
                    SET @tlen = @tlen / 2;
                IF @ttype IN (N~nvarchar~, N~nchar~, N~varchar~, N~char~) AND @tlen <> -1
                    SET @sql = N~SELECT @n = COUNT(*) FROM ~ + @stage + N~ WHERE LEN(~ + QUOTENAME(@col) + N~ + N~~.~~) - 1 > ~
                             + CAST(@tlen AS nvarchar(20)) + N~;~;
                ELSE IF @ttype IN (N~bigint~, N~int~, N~smallint~, N~tinyint~) AND @stype IN (N~float~, N~real~, N~decimal~, N~numeric~)
                    SET @sql = N~SELECT @n = COUNT(*) FROM ~ + @stage + N~ WHERE ~ + QUOTENAME(@col) + N~ <> FLOOR(~ + QUOTENAME(@col) + N~);~;
                IF @sql IS NOT NULL
                BEGIN
                    EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
                    IF @n > 0
                        INSERT INTO #problems VALUES (@target, N~column ~ + @col + N~: ~ + CAST(@n AS nvarchar(20))
                            + CASE WHEN @ttype IN (N~bigint~, N~int~, N~smallint~, N~tinyint~)
                                   THEN N~ value(s) have a fraction, and this database keeps whole numbers only (~ + @ttype + N~)~
                                   ELSE N~ value(s) are longer than this database allows (~ + @ttype + N~(~
                                        + CAST(@tlen AS nvarchar(20)) + N~))~ END);
                END
                FETCH NEXT FROM fit_cur INTO @col, @ttype, @tlen, @stype;
            END
            CLOSE fit_cur;
            DEALLOCATE fit_cur;
        END
    END
    FETCH NEXT FROM plan_cur INTO @ord, @target, @stage, @rows_in_file, @placeholder;
END
CLOSE plan_cur;
DEALLOCATE plan_cur;

IF EXISTS (SELECT 1 FROM #problems)
BEGIN
    SELECT N~NOTHING WAS WRITTEN~ AS result, target AS [table], problem FROM #problems ORDER BY target;
    IF EXISTS (SELECT 1 FROM #notes)
        SELECT target AS [table], note FROM #notes ORDER BY target;
    RAISERROR(N~This database cannot take this file (see the Results tab). Nothing was written.~, 16, 1);
    RETURN;
END

-- 2. Write ----------------------------------------------------------------------
BEGIN TRANSACTION;

INSERT INTO #loaded (ord, target, rows_in_file, placeholders_removed, rows_loaded)
SELECT ord, target, rows_in_file, 0, 0 FROM #plan;

-- The API~s placeholder rows go first, children before parents.
DECLARE ph_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, placeholder FROM #plan
     WHERE placeholder IS NOT NULL AND OBJECT_ID(N~dbo.~ + QUOTENAME(target), N~U~) IS NOT NULL
     ORDER BY ord DESC;
OPEN ph_cur;
FETCH NEXT FROM ph_cur INTO @ord, @target, @placeholder;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N~DELETE FROM dbo.~ + QUOTENAME(@target) + N~ WHERE ~ + @placeholder + N~; SET @n = @@ROWCOUNT;~;
    EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
    UPDATE #loaded SET placeholders_removed = @n WHERE ord = @ord;
    FETCH NEXT FROM ph_cur INTO @ord, @target, @placeholder;
END
CLOSE ph_cur;
DEALLOCATE ph_cur;

-- Then the rows, parents before children, each keeping the id it has in the source.
DECLARE load_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT ord, target, stage FROM #plan
     WHERE stage IS NOT NULL AND OBJECT_ID(N~dbo.~ + QUOTENAME(target), N~U~) IS NOT NULL
     ORDER BY ord;
OPEN load_cur;
FETCH NEXT FROM load_cur INTO @ord, @target, @stage;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @collist = STUFF((SELECT N~, ~ + QUOTENAME(col) FROM #cols
                           WHERE ord = @ord AND load_it = 1 ORDER BY colid FOR XML PATH(~~)), 1, 2, N~~);
    SET @identity = CASE WHEN EXISTS (
            SELECT 1 FROM sys.columns t
              JOIN #cols c ON c.ord = @ord AND c.load_it = 1 AND c.col = t.name COLLATE DATABASE_DEFAULT
             WHERE t.object_id = OBJECT_ID(N~dbo.~ + QUOTENAME(@target)) AND t.is_identity = 1) THEN 1 ELSE 0 END;
    SET @sql = CASE WHEN @identity = 1 THEN N~SET IDENTITY_INSERT dbo.~ + QUOTENAME(@target) + N~ ON; ~ ELSE N~~ END
             + N~INSERT INTO dbo.~ + QUOTENAME(@target) + N~ (~ + @collist + N~) SELECT ~ + @collist
             + N~ FROM ~ + @stage + N~; SET @n = @@ROWCOUNT; ~
             + CASE WHEN @identity = 1 THEN N~SET IDENTITY_INSERT dbo.~ + QUOTENAME(@target) + N~ OFF;~ ELSE N~~ END;
    EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
    UPDATE #loaded SET rows_loaded = @n WHERE ord = @ord;
    IF @n <> (SELECT rows_in_file FROM #loaded WHERE ord = @ord)
        INSERT INTO #problems VALUES (@target, N~loaded ~ + CAST(@n AS nvarchar(20)) + N~ rows, expected ~
            + CAST((SELECT rows_in_file FROM #loaded WHERE ord = @ord) AS nvarchar(20)));
    FETCH NEXT FROM load_cur INTO @ord, @target, @stage;
END
CLOSE load_cur;
DEALLOCATE load_cur;

-- 3. Every reference still lands -------------------------------------------------
-- Counted the same way in the source when this file was made. A reference that was
-- already dangling there is carried as it is; one that dangles only here is a failure.
DECLARE ref_cur CURSOR LOCAL FAST_FORWARD FOR
    SELECT child, col, parent, pcol FROM #refs;
OPEN ref_cur;
FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
WHILE @@FETCH_STATUS = 0
BEGIN
    IF COL_LENGTH(N~dbo.~ + QUOTENAME(@child), @col) IS NOT NULL AND COL_LENGTH(N~dbo.~ + QUOTENAME(@parent), @pcol) IS NOT NULL
    BEGIN
        SET @sql = N~SELECT @n = COUNT(*) FROM dbo.~ + QUOTENAME(@child) + N~ c WHERE c.~ + QUOTENAME(@col)
                 + N~ <> 0 AND NOT EXISTS (SELECT 1 FROM dbo.~ + QUOTENAME(@parent) + N~ p WHERE p.~ + QUOTENAME(@pcol)
                 + N~ = c.~ + QUOTENAME(@col) + N~);~;
        EXEC sp_executesql @sql, N~@n int OUTPUT~, @n = @n OUTPUT;
        UPDATE #refs SET dangling_here = @n WHERE child = @child AND col = @col AND parent = @parent AND pcol = @pcol;
    END
    FETCH NEXT FROM ref_cur INTO @child, @col, @parent, @pcol;
END
CLOSE ref_cur;
DEALLOCATE ref_cur;

INSERT INTO #problems
SELECT child, N~column ~ + col + N~: ~ + CAST(dangling_here AS nvarchar(20)) + N~ row(s) point at a ~ + parent
     + N~ row that is not there (the source had ~ + CAST(dangling_in_source AS nvarchar(20))
     + N~). Run the earlier files of this batch first.~
  FROM #refs WHERE dangling_here > dangling_in_source;

IF EXISTS (SELECT 1 FROM #problems)
BEGIN
    ROLLBACK TRANSACTION;
    SELECT N~NOTHING WAS WRITTEN~ AS result, target AS [table], problem FROM #problems ORDER BY target;
    RAISERROR(N~The loaded rows failed a check (see the Results tab). Everything was rolled back.~, 16, 1);
    RETURN;
END

-- 4. What happened ----------------------------------------------------------------
SELECT target AS [table], rows_in_file, placeholders_removed AS api_placeholders_removed, rows_loaded
  FROM #loaded ORDER BY ord;

IF EXISTS (SELECT 1 FROM #refs WHERE dangling_in_source > 0 OR dangling_here > 0)
    SELECT child + N~.~ + col AS [reference], parent + N~.~ + pcol AS points_at,
           dangling_in_source, dangling_here, N~carried as it was in the source~ AS note
      FROM #refs WHERE dangling_in_source > 0 OR dangling_here > 0 ORDER BY child, col;

IF EXISTS (SELECT 1 FROM #notes)
    SELECT target AS [table], note FROM #notes ORDER BY target;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT N~committed~;
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT N~rolled back - set @commit = 1 to write~;
END';

INSERT INTO #out (txt) VALUES (N''), (REPLACE(@engine, N'~', NCHAR(39))), (N'');

INSERT INTO #out (txt)
SELECT N'IF OBJECT_ID(''tempdb..#m_' + tbl + N''') IS NOT NULL DROP TABLE #m_' + tbl + N';'
  FROM #pick WHERE delete_only = 0 ORDER BY ord;

INSERT INTO #out (txt) VALUES
 (N'IF OBJECT_ID(''tempdb..#plan'')     IS NOT NULL DROP TABLE #plan;'),
 (N'IF OBJECT_ID(''tempdb..#refs'')     IS NOT NULL DROP TABLE #refs;'),
 (N'IF OBJECT_ID(''tempdb..#problems'') IS NOT NULL DROP TABLE #problems;'),
 (N'IF OBJECT_ID(''tempdb..#notes'')    IS NOT NULL DROP TABLE #notes;'),
 (N'IF OBJECT_ID(''tempdb..#cols'')     IS NOT NULL DROP TABLE #cols;'),
 (N'IF OBJECT_ID(''tempdb..#loaded'')   IS NOT NULL DROP TABLE #loaded;');

-- sqlcmd -y 0 cuts a single value at 1 MB. Refuse rather than write a broken line.
IF EXISTS (SELECT 1 FROM #out WHERE LEN(txt) > 500000)
BEGIN
    RAISERROR('A generated value is over 500,000 characters - sqlcmd -y 0 would cut it.', 16, 1);
    RETURN;
END

SELECT txt FROM #out ORDER BY n;
