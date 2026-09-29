-- STEP 01 of 04 - put rehearsal's Setup back as it was before the 2026-09-25 wipe
-- Run the numbered steps in order, each on its own, all on Lightspeed_ERP_deploy_rehearsal:
--   01  this file
--   02  02_add_controller_prerequisites_2026-09-28.sql
--   03  ..\2026-09-25_01_controllers_and_project_quote\05_load_controllers_items_2026-09-23.sql   (@commit = 1)
--   04  ..\2026-09-25_01_controllers_and_project_quote\07_load_controllers_bom_2026-09-23.sql     (@commit = 1, @round_up_fractional_qty = 1)
-- 01 MUST come first: it makes these 14 tables equal to the pre-wipe copy, so it would take
-- out any controller item loaded before it.
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- What happened
-- =============
-- "03 migrate_setup_data_2026-09-08.sql" (a loader for an EMPTY database: it DELETEs 34 Setup
-- tables and reloads a 2026-09-08 snapshot) was run on rehearsal at 2026-09-25 14:56 and again
-- 2026-09-28 10:00. Rehearsal lost the 2,440 Calpeda pumps and their FLA/VOLTAGE, the 226-account
-- chart, 16 of its 17 users and more, and 11 foreign keys were left switched off.
--
-- Lightspeed_rehearsal_before_wipe is a point-in-time restore of rehearsal (the 2026-09-09 full
-- backup + a copy-only log backup taken 2026-09-28) stopped just before that first run
-- (STOPBEFOREMARK lsn:424000002507200002, 2026-09-25 14:56:13.793). A row-by-row comparison of
-- the 34 tables found 14 that differ; the other 20 came back identical. Nothing created on
-- rehearsal since then points at a row this removes (checked 2026-09-28).
--
-- What it does, in one transaction, on those 14 tables only, matching rows on id:
--   rows the wipe added are deleted, rows it changed are updated, rows it deleted are inserted
--   with their original ids. Then every table is compared again and the run refuses to commit
--   if any still differs. Finally the foreign keys that were on before the wipe and are off
--   now are switched back on WITH CHECK (the check counted 0 offending rows).
--
-- Leaves alone: every other table - quotes, orders, history, audit rows made since 09-25.
-- Users: the one Admin the snapshot left has id 1; the pre-wipe Admin (same employee id
-- IT-WD-1) has id 2, so anyone logged in to 3001 must log in again afterwards.
--
-- HOW TO RUN: sqlcmd -S <server> -d Lightspeed_ERP_deploy_rehearsal -b -I -i <this file>
-- (or SSMS - CHECK THE DATABASE DROPDOWN). After a commit, restart 3001 or clear its Redis keys.
SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

IF DB_ID(N'Lightspeed_rehearsal_before_wipe') IS NULL
BEGIN
    RAISERROR(N'Lightspeed_rehearsal_before_wipe (the point-in-time copy) is not on this server.', 16, 1);
    RETURN;
END

IF DB_NAME() <> N'Lightspeed_ERP_deploy_rehearsal'
BEGIN
    RAISERROR(N'Run this on Lightspeed_ERP_deploy_rehearsal only.', 16, 1);
    RETURN;
END

DECLARE @tables TABLE (ord int IDENTITY PRIMARY KEY, tbl sysname NOT NULL);
INSERT INTO @tables (tbl) VALUES
    (N'tbl_inv_warehouse_usetype'), (N'tbl_position'), (N'tbl_position_access'),
    (N'tbl_setup_bpi_industries'), (N'tbl_setup_chart_class'), (N'tbl_setup_chart_of_accounts'),
    (N'tbl_setup_item'), (N'tbl_setup_item_additional_specs'), (N'tbl_setup_item_inventory'),
    (N'tbl_setup_item_material'), (N'tbl_setup_item_specs'), (N'tbl_setup_item_specs_template'),
    (N'tbl_setup_item_trade_type'), (N'tbl_setup_users');

DECLARE @result TABLE (tbl sysname, deleted int, updated int, inserted int, rows_now int, still_differs int);

DECLARE @tbl sysname, @cols nvarchar(max), @cmp_live nvarchar(max), @cmp_src nvarchar(max), @set nvarchar(max),
        @sql nvarchar(max), @d int, @u int, @i int, @n int, @x int;

BEGIN TRANSACTION;

DECLARE tcur CURSOR LOCAL FAST_FORWARD FOR SELECT tbl FROM @tables ORDER BY ord;
OPEN tcur;
FETCH NEXT FROM tcur INTO @tbl;
WHILE @@FETCH_STATUS = 0
BEGIN
    SELECT @cols = STUFF((SELECT N',' + QUOTENAME(c.name) FROM sys.columns c
                           WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@tbl)) AND c.is_computed = 0
                           ORDER BY c.column_id FOR XML PATH('')), 1, 1, N''),
           @set  = STUFF((SELECT N',' + QUOTENAME(c.name) + N'=s.' + QUOTENAME(c.name) FROM sys.columns c
                           WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@tbl)) AND c.is_computed = 0 AND c.is_identity = 0
                           ORDER BY c.column_id FOR XML PATH('')), 1, 1, N''),
           @cmp_live = STUFF((SELECT N',t.' + QUOTENAME(c.name) FROM sys.columns c JOIN sys.types ty ON ty.user_type_id = c.user_type_id
                           WHERE c.object_id = OBJECT_ID(N'dbo.' + QUOTENAME(@tbl)) AND c.is_computed = 0
                             AND ty.name NOT IN (N'text', N'ntext', N'image', N'xml')
                           ORDER BY c.column_id FOR XML PATH('')), 1, 1, N'');
    SET @cmp_src = REPLACE(@cmp_live, N't.', N's.');

    SET @sql = N'DELETE t FROM dbo.' + QUOTENAME(@tbl) + N' t
                  WHERE NOT EXISTS (SELECT 1 FROM Lightspeed_rehearsal_before_wipe.dbo.' + QUOTENAME(@tbl) + N' s WHERE s.id = t.id);
                 SET @d = @@ROWCOUNT;
                 UPDATE t SET ' + @set + N'
                   FROM dbo.' + QUOTENAME(@tbl) + N' t
                   JOIN Lightspeed_rehearsal_before_wipe.dbo.' + QUOTENAME(@tbl) + N' s ON s.id = t.id
                  WHERE EXISTS (SELECT ' + @cmp_live + N' EXCEPT SELECT ' + @cmp_src + N');
                 SET @u = @@ROWCOUNT;
                 SET IDENTITY_INSERT dbo.' + QUOTENAME(@tbl) + N' ON;
                 INSERT INTO dbo.' + QUOTENAME(@tbl) + N' (' + @cols + N')
                 SELECT ' + @cols + N' FROM Lightspeed_rehearsal_before_wipe.dbo.' + QUOTENAME(@tbl) + N' s
                  WHERE NOT EXISTS (SELECT 1 FROM dbo.' + QUOTENAME(@tbl) + N' t WHERE t.id = s.id);
                 SET @i = @@ROWCOUNT;
                 SET IDENTITY_INSERT dbo.' + QUOTENAME(@tbl) + N' OFF;
                 SELECT @n = COUNT(*) FROM dbo.' + QUOTENAME(@tbl) + N';
                 SELECT @x = COUNT(*) FROM (
                     SELECT ' + REPLACE(@cmp_live, N't.', N'') + N' FROM dbo.' + QUOTENAME(@tbl) + N'
                     EXCEPT SELECT ' + REPLACE(@cmp_live, N't.', N'') + N' FROM Lightspeed_rehearsal_before_wipe.dbo.' + QUOTENAME(@tbl) + N'
                     UNION ALL
                     SELECT ' + REPLACE(@cmp_live, N't.', N'') + N' FROM Lightspeed_rehearsal_before_wipe.dbo.' + QUOTENAME(@tbl) + N'
                     EXCEPT SELECT ' + REPLACE(@cmp_live, N't.', N'') + N' FROM dbo.' + QUOTENAME(@tbl) + N') z;';
    EXEC sp_executesql @sql, N'@d int OUTPUT, @u int OUTPUT, @i int OUTPUT, @n int OUTPUT, @x int OUTPUT',
         @d OUTPUT, @u OUTPUT, @i OUTPUT, @n OUTPUT, @x OUTPUT;

    INSERT INTO @result VALUES (@tbl, @d, @u, @i, @n, @x);
    FETCH NEXT FROM tcur INTO @tbl;
END
CLOSE tcur;
DEALLOCATE tcur;

SELECT tbl, deleted, updated, inserted, rows_now, still_differs FROM @result;

-- Rows that would break each foreign key the wipe left switched off (all single-column).
DECLARE @fk TABLE (fk sysname, child sysname, orphans int);
DECLARE @fkname sysname, @child sysname, @ccol sysname, @parent sysname, @pcol sysname, @o int;
DECLARE fcur CURSOR LOCAL FAST_FORWARD FOR
    SELECT fk.name, OBJECT_NAME(fk.parent_object_id), pc.name, OBJECT_NAME(fk.referenced_object_id), rc.name
      FROM sys.foreign_keys fk
      JOIN sys.foreign_key_columns fkc ON fkc.constraint_object_id = fk.object_id
      JOIN sys.columns pc ON pc.object_id = fkc.parent_object_id AND pc.column_id = fkc.parent_column_id
      JOIN sys.columns rc ON rc.object_id = fkc.referenced_object_id AND rc.column_id = fkc.referenced_column_id
     WHERE fk.is_disabled = 1;
OPEN fcur;
FETCH NEXT FROM fcur INTO @fkname, @child, @ccol, @parent, @pcol;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N'SELECT @o = COUNT(*) FROM dbo.' + QUOTENAME(@child) + N' c WHERE c.' + QUOTENAME(@ccol) + N' IS NOT NULL
                   AND NOT EXISTS (SELECT 1 FROM dbo.' + QUOTENAME(@parent) + N' p WHERE p.' + QUOTENAME(@pcol) + N' = c.' + QUOTENAME(@ccol) + N');';
    EXEC sp_executesql @sql, N'@o int OUTPUT', @o OUTPUT;
    INSERT INTO @fk VALUES (@fkname, @child, @o);
    FETCH NEXT FROM fcur INTO @fkname, @child, @ccol, @parent, @pcol;
END
CLOSE fcur;
DEALLOCATE fcur;

SELECT fk, child, orphans FROM @fk ORDER BY fk;

IF EXISTS (SELECT 1 FROM @result WHERE still_differs <> 0)
BEGIN
    ROLLBACK TRANSACTION;
    RAISERROR(N'A table still differs from the pre-wipe copy - rolled back, nothing changed.', 16, 1);
    RETURN;
END

IF EXISTS (SELECT 1 FROM @fk WHERE orphans <> 0)
BEGIN
    ROLLBACK TRANSACTION;
    RAISERROR(N'A foreign key has offending rows (listed above) - rolled back, nothing changed.', 16, 1);
    RETURN;
END

-- Switch back on the foreign keys that were on before the wipe and are off now.
DECLARE @reenabled TABLE (fk sysname, child sysname);
DECLARE rcur CURSOR LOCAL FAST_FORWARD FOR
    SELECT n.name, OBJECT_NAME(n.parent_object_id)
      FROM sys.foreign_keys n
      JOIN Lightspeed_rehearsal_before_wipe.sys.foreign_keys b ON b.name = n.name
     WHERE n.is_disabled = 1 AND b.is_disabled = 0;
OPEN rcur;
FETCH NEXT FROM rcur INTO @fkname, @child;
WHILE @@FETCH_STATUS = 0
BEGIN
    SET @sql = N'ALTER TABLE dbo.' + QUOTENAME(@child) + N' WITH CHECK CHECK CONSTRAINT ' + QUOTENAME(@fkname) + N';';
    EXEC sp_executesql @sql;
    INSERT INTO @reenabled VALUES (@fkname, @child);
    FETCH NEXT FROM rcur INTO @fkname, @child;
END
CLOSE rcur;
DEALLOCATE rcur;

SELECT fk AS foreign_key_switched_back_on, child FROM @reenabled ORDER BY fk;

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT N'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT N'trial run - rolled back, nothing changed';
END
