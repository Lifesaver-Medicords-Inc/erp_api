-- Generator: emits a portable INSERT script for every tbl_setup_* table holding
-- rows. Run against the SOURCE database; pipe stdout to the .sql you will carry
-- to the 2012 server.
-- FOR XML PATH(...).value() is an XML data type method, which SQL Server
-- refuses unless QUOTED_IDENTIFIER is ON. sqlcmd turns it OFF by default.
SET QUOTED_IDENTIFIER ON;
SET NOCOUNT ON;

DECLARE @tbl        sysname,
        @cols       nvarchar(max),
        @vals       nvarchar(max),
        @sql        nvarchar(max),
        @hasIdent   bit,
        @rows       int;

-- Columns that exist only in the legacy schema. The Go models no longer declare
-- them, so a database built by AutoMigrate has no such column and an INSERT
-- naming it fails with "Invalid column name". Verified empty on the source
-- before exclusion - tbl_setup_item_specs.title and .value are NULL on all 24
-- rows - so nothing is dropped by leaving them out.
--
-- Re-check this list if the source schema changes: the comparison that produced
-- it is a NOT EXISTS between the source's columns and a freshly migrated
-- database's columns, per table.
DECLARE @excluded TABLE (tbl sysname, col sysname);
INSERT INTO @excluded (tbl, col) VALUES
    ('tbl_setup_item_specs', 'title'),
    ('tbl_setup_item_specs', 'value');

PRINT '-- ============================================================';
PRINT '-- Lightspeed ERP - setup data';
PRINT '-- Generated from Lightspeed_ERP_deploy_rehearsal';
PRINT '-- Target: SQL Server 2012.';
PRINT '--';
PRINT '-- HOW TO RUN';
PRINT '--   sqlcmd -S <server> -E -d <new database> -b -I -i <this file>';
PRINT '--   The -I matters: QUOTED_IDENTIFIER must be ON.';
PRINT '--';
PRINT '-- Run AFTER the API has started once against the new database, so';
PRINT '-- AutoMigrate has created the tables. This loads data only - it creates';
PRINT '-- nothing. Each table is DELETEd before loading, so re-running is safe';
PRINT '-- and leaves the same rows rather than duplicates.';
PRINT '--';
PRINT '-- Foreign keys are suspended for the load and restored WITH CHECK at the';
PRINT '-- end, so table order does not matter and any inconsistency is reported.';
PRINT '--';
PRINT '-- REGENERATE with sql/snippets/_generate_setup_data.sql:';
PRINT '--   sqlcmd -S <src> -E -d <src db> -b -I -w 65535 -y 0 -u -i _generate_setup_data.sql -o <out>';
PRINT '--   -y 0 is required (sqlcmd truncates at 256 chars without it)';
PRINT '--   -u  is required (data contains non-ASCII characters)';
PRINT '-- ============================================================';
PRINT 'SET NOCOUNT ON;';
PRINT 'GO';
PRINT '';

-- Foreign keys are switched off around the load so table order does not matter,
-- then switched back on WITH CHECK so anything inconsistent is reported rather
-- than left untrusted.
PRINT '-- Suspend FK checking for the load';
-- Emitted a row at a time via SELECT, not accumulated into one string and
-- PRINTed: PRINT truncates nvarchar at 4000 characters, and this list is longer
-- than that - which silently cut a statement in half and produced a file that
-- would not parse.
SELECT CAST('' AS nvarchar(max))
     + 'IF OBJECT_ID(''dbo.' + t.name + ''') IS NOT NULL ALTER TABLE [dbo].[' + t.name + '] NOCHECK CONSTRAINT ALL;'
FROM   sys.tables t
WHERE  EXISTS (SELECT 1 FROM sys.foreign_keys f WHERE f.parent_object_id = t.object_id)
ORDER  BY t.name;
PRINT 'GO';
PRINT '';

DECLARE tables CURSOR LOCAL FAST_FORWARD FOR
    SELECT t.name
    FROM   sys.tables t
    JOIN   sys.partitions p ON p.object_id = t.object_id AND p.index_id IN (0,1)
    -- Everything named tbl_setup_*, plus four tables that are setup data in
    -- everything but name:
    --
    --   tbl_position
    --     Not optional. tbl_setup_users has an FK to tbl_position.id
    --     (fk_tbl_position_users), so loading users without positions leaves the
    --     constraint unsatisfiable and the closing WITH CHECK fails. It is the
    --     only such dependency - verified by listing every FK whose child is a
    --     setup table and whose parent is not.
    --
    --   tbl_company
    --     Company Setup (spec 4.5.6). Holds vat_rate_percent and
    --     markup_multiplier_price, which quotation and invoice maths read.
    --
    --   tbl_trans_sales_project_template(_child)
    --     Project Quotation templates. Named tbl_trans_* but they are
    --     configuration, not transactions. Note initializers.
    --     SeedDefaultProjectQuotationTemplate() creates a placeholder on a
    --     genuinely empty database; the DELETE before each table's inserts
    --     clears that placeholder, which is the intended outcome.
    WHERE  (t.name LIKE 'tbl_setup%'
            OR t.name IN ('tbl_position',
                          'tbl_company',
                          'tbl_trans_sales_project_template',
                          'tbl_trans_sales_project_template_child'))
    GROUP  BY t.name
    HAVING SUM(p.rows) > 0
    ORDER  BY t.name;

OPEN tables;
FETCH NEXT FROM tables INTO @tbl;

WHILE @@FETCH_STATUS = 0
BEGIN
    -- Bracketed column list: several of these tables use reserved words
    -- ("group" on tbl_setup_chart_of_accounts, for one).
    SELECT @cols = STUFF((
        SELECT ', [' + c.name + ']'
        FROM   sys.columns c
        WHERE  c.object_id = OBJECT_ID('dbo.' + @tbl)
        AND NOT EXISTS (SELECT 1 FROM @excluded e WHERE e.tbl = @tbl AND e.col = c.name)
        ORDER  BY c.column_id
        FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 2, '');

    -- One expression per column, rendering a SQL literal. Only five types occur
    -- across these tables (bigint, int, bit, float, nvarchar), so this covers
    -- them all; anything else would surface as an error rather than bad data.
    SELECT @vals = STUFF((
        SELECT ' + '', '' + ' +
               CASE
                 -- N'...' , not '...': six item_model values and six long_description
                 -- values hold characters outside plain ASCII. Without the N prefix
                 -- the literal is read in the database's codepage and anything not
                 -- representable there is silently replaced.
                 --
                 -- CHAR(39) rather than stacked doubled quotes - at this nesting depth
                 -- (a generator emitting SQL that emits SQL) literal quotes stop being
                 -- readable and start being guesswork.
                 WHEN t.name = 'nvarchar'
                   THEN 'ISNULL(''N'' + CHAR(39) + REPLACE(CAST([' + c.name + '] AS nvarchar(max)), CHAR(39), CHAR(39)+CHAR(39)) + CHAR(39), ''NULL'')'
                 -- varchar(10), not varchar(1): ISNULL takes its return type from
                 -- the FIRST argument, so a narrow cast silently truncates the
                 -- 'NULL' replacement to 'N' and emits an unquoted identifier.
                 WHEN t.name = 'bit'
                   THEN 'ISNULL(CAST(CAST([' + c.name + '] AS int) AS varchar(10)), ''NULL'')'
                 WHEN t.name = 'float'
                   THEN 'ISNULL(CONVERT(varchar(40), [' + c.name + '], 2), ''NULL'')'
                 ELSE 'ISNULL(CAST([' + c.name + '] AS varchar(30)), ''NULL'')'
               END
        FROM   sys.columns c
        JOIN   sys.types t ON t.user_type_id = c.user_type_id
        WHERE  c.object_id = OBJECT_ID('dbo.' + @tbl)
        AND NOT EXISTS (SELECT 1 FROM @excluded e WHERE e.tbl = @tbl AND e.col = c.name)
        ORDER  BY c.column_id
        FOR XML PATH(''), TYPE).value('.', 'nvarchar(max)'), 1, 10, '');

    SELECT @hasIdent = CASE WHEN EXISTS (
        SELECT 1 FROM sys.identity_columns WHERE object_id = OBJECT_ID('dbo.' + @tbl)
    ) THEN 1 ELSE 0 END;

    PRINT '-- ---------- ' + @tbl + ' ----------';
    PRINT 'DELETE FROM [dbo].[' + @tbl + '];';
    IF @hasIdent = 1
        PRINT 'SET IDENTITY_INSERT [dbo].[' + @tbl + '] ON;';

    -- The leading CAST forces the whole concatenation to nvarchar(max). Without
    -- it SQL Server infers a shorter type, and sqlcmd then wraps the output at
    -- that width - splitting long INSERTs mid-literal and producing a file that
    -- will not parse ("Incorrect syntax near 'META'").
    SET @sql = 'SELECT CAST('''' AS nvarchar(max)) + ''INSERT INTO [dbo].[' + @tbl + '] (' + @cols + ') VALUES ('' + '
             + @vals + ' + '');'' FROM [dbo].[' + @tbl + '];';
    EXEC sp_executesql @sql;

    IF @hasIdent = 1
        PRINT 'SET IDENTITY_INSERT [dbo].[' + @tbl + '] OFF;';
    PRINT 'GO';
    PRINT '';

    FETCH NEXT FROM tables INTO @tbl;
END

CLOSE tables;
DEALLOCATE tables;

PRINT '-- Re-enable FK checking. WITH CHECK validates the loaded rows, so a';
PRINT '-- violation is reported here rather than leaving an untrusted constraint.';
SELECT CAST('' AS nvarchar(max))
     + 'IF OBJECT_ID(''dbo.' + t.name + ''') IS NOT NULL ALTER TABLE [dbo].[' + t.name + '] WITH CHECK CHECK CONSTRAINT ALL;'
FROM   sys.tables t
WHERE  EXISTS (SELECT 1 FROM sys.foreign_keys f WHERE f.parent_object_id = t.object_id)
ORDER  BY t.name;
PRINT 'GO';
