-- QERP catalogue corrections the controller BOMs depend on
-- ========================================================
-- From the 'QERP Actions' sheet of 'Controllers_Data_Entry_(9_22_26).xlsx'. It lists the
-- items already in the catalogue that the controller BOMs point at under a different brand
-- or item name, plus a few the workbook supersedes outright.
--
-- Three groups, each behind its own switch, all OFF except the edits:
--
--   @apply_edits (default 1)
--       15 field corrections. Five of them the BOM load NEEDS: the BOM names
--       ANLY/OMR/WIX/MAC/CHS where QERP has the same item under 'No Brand', so without
--       them those components resolve to nothing and load_controllers_bom_*.sql refuses to
--       run. The rest move the nine CONTROL BOX items from item name CTRL to ME (metal
--       enclosure), which is what they are - CTRL is now the controllers themselves.
--
--   @apply_deletes (default 0)
--       6 items the workbook says are superseded: five unbranded control voltage
--       transformers replaced by the Sakura Denso / Yokoyama pair of the same spec, and
--       'SQUARE D PRESSURE SWITCH' replaced by one carrying its psi range. DESTRUCTIVE, so
--       it is off. An item that is referenced ANYWHERE - any transaction table with an
--       item_id, its own BOM, a BOM that uses it - is left alone and reported instead of
--       deleted. Turning this on only ever removes items nothing points at.
--
--   @apply_optional_brands (default 0)
--       46 housekeeping fills: QERP items whose name already says the brand ('FUJI
--       MAGNETIC CONTACTOR ...') while the brand field says No Brand. Nothing depends on
--       these; they are here so the list is complete.
--
-- Every row is matched on ITEM MODEL + the item name and brand the workbook recorded, so a
-- row that has already been corrected, or that never looked like this on your database,
-- is skipped and reported rather than matched by model alone.
--
-- HOW TO RUN: sqlcmd -S <server> -d <database> -b -i <this file>  (or open it in SSMS).
-- With @commit = 0 everything is done and reported, then rolled back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;                 -- 1 = keep the changes, 0 = trial run
DECLARE @apply_edits bit = 1;            -- the corrections the controller BOMs depend on
DECLARE @apply_deletes bit = 0;          -- superseded items - destructive, see above
DECLARE @apply_optional_brands bit = 0;  -- brand housekeeping, nothing depends on it

IF OBJECT_ID('tempdb..#action') IS NOT NULL DROP TABLE #action;
IF OBJECT_ID('tempdb..#done')   IS NOT NULL DROP TABLE #done;
IF OBJECT_ID('tempdb..#refs')   IS NOT NULL DROP TABLE #refs;

CREATE TABLE #action (
    row_no     int NOT NULL PRIMARY KEY,
    kind       nvarchar(20)  NOT NULL,   -- EDIT / DELETE / OPTIONAL
    item_model nvarchar(400) NOT NULL,
    name_code  nvarchar(50)  NOT NULL,   -- the item name the workbook saw
    brand_code nvarchar(50)  NOT NULL,   -- the brand the workbook saw
    new_name   nvarchar(50)  NOT NULL,   -- '' = leave it
    new_brand  nvarchar(50)  NOT NULL,   -- '' = leave it
    why        nvarchar(400) NOT NULL,
    item_id    bigint NULL
);
CREATE TABLE #done (row_no int PRIMARY KEY, outcome nvarchar(60) NOT NULL);
CREATE TABLE #refs (item_id bigint, referenced_by nvarchar(300));

-- The sheet ---------------------------------------------------------------------
INSERT INTO #action (row_no, kind, item_model, name_code, brand_code, new_name, new_brand, why) VALUES
    (1, N'EDIT', N'ANLY STAR DELTA TIMER TRD-N 220V W/ SOCKET', N'ELEC', N'NB', N'', N'ANLY', N'Brand NB -> ANLY'),
    (2, N'EDIT', N'CABLE TIE 4"', N'HDW', N'NB', N'', N'CHS', N'Brand NB -> CHS'),
    (3, N'EDIT', N'CONTROL BOX 16" X 14" X 8" N3R', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (4, N'EDIT', N'CONTROL BOX 20" X 17" X 9" N3R', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (5, N'EDIT', N'CONTROL BOX 20" X 20" X 9" N3R', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (6, N'EDIT', N'CONTROL BOX 25" X 20" X 9" NR3', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (7, N'EDIT', N'CONTROL BOX 28" X 22" X 9" N3R', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (8, N'EDIT', N'CONTROL BOX 28" X 24" X 10" N3R', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (9, N'EDIT', N'CONTROL BOX 36" X 26" X 14" N3R W/ VIEWING GLASS, WRINKLED BEIGE', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (10, N'EDIT', N'CONTROL BOX 400MM X 300MM X 200MM', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (11, N'EDIT', N'CONTROL BOX 600MM X 500MM X 200MM', N'CTRL', N'NB', N'ME', N'', N'Item Name CTRL -> ME'),
    (12, N'EDIT', N'MAC 3 FLOAT SWITCH', N'ELEC', N'NB', N'FS', N'MAC', N'Item Name ELEC -> FS; Brand NB -> MAC'),
    (13, N'EDIT', N'OMRON FLOATLESS RELAY 61FG (DOWN)', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR'),
    (14, N'EDIT', N'OMRON FLOATLESS RELAY 61FG1', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR'),
    (15, N'EDIT', N'TIE MOUNT 1" X 1"', N'ELEC', N'NB', N'', N'WIX', N'Brand NB -> WIX'),
    (16, N'DELETE', N'CONTROL VOLTAGE TRANSFORMER 100VA, 480V/220V', N'ELEC', N'NB', N'', N'', N'unbranded - replaced by the Sakura Denso + Yokoyama pair of the same spec'),
    (17, N'DELETE', N'CONTROL VOLTAGE TRANSFORMER 200VA 480/220V', N'ELEC', N'NB', N'', N'', N'unbranded - replaced by the Sakura Denso + Yokoyama pair of the same spec'),
    (18, N'DELETE', N'CONTROL VOLTAGE TRANSFORMER 200VA, 1PH, 380/400/440V / 220V', N'ELEC', N'NB', N'', N'', N'unbranded - replaced by the Sakura Denso + Yokoyama pair (the spec the BOMs use)'),
    (19, N'DELETE', N'CONTROL VOLTAGE TRANSFORMER 300VA , 400V/480V / 220V', N'ELEC', N'NB', N'', N'', N'unbranded - replaced by the Sakura Denso + Yokoyama pair of the same spec'),
    (20, N'DELETE', N'CONTROL VOLTAGE TRANSFORMER 300VA 380/400/440V / 220V', N'ELEC', N'NB', N'', N'', N'unbranded - replaced by the Sakura Denso + Yokoyama pair of the same spec'),
    (21, N'DELETE', N'SQUARE D PRESSURE SWITCH', N'ELEC', N'NB', N'', N'', N'replaced by PS/SQD PRESSURE SWITCH 20-40PSI, which carries the psi range'),
    (22, N'OPTIONAL', N'ANLY TIMER 0-30 SEC. AH3- 220V W/ SOCKET', N'ELEC', N'NB', N'', N'ANLY', N'Brand NB -> ANLY (housekeeping; nothing depends on it)'),
    (23, N'OPTIONAL', N'ANLY U/O VOLTAGE RELAY AVR165 380V 11PINS', N'ELEC', N'NB', N'', N'ANLY', N'Brand NB -> ANLY (housekeeping; nothing depends on it)'),
    (24, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 100A, 3P, 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (25, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 15A , 3P, 440V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (26, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 15A 2POLE 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (27, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 15A, 3P, 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (28, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 20A , 2P, 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (29, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 20A 2POLE 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (30, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 30A, 3P, 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (31, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 40A 3P 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (32, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 40A, 2P, 440V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (33, N'OPTIONAL', N'FUJI CIRCUIT BREAKER 50A, 3P, 500V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (34, N'OPTIONAL', N'FUJI HOME PUMP', N'PMP', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (35, N'OPTIONAL', N'FUJI MAGNETIC CONTACTOR SC-05, 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (36, N'OPTIONAL', N'FUJI MAGNETIC CONTACTOR SC-5-1, 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (37, N'OPTIONAL', N'FUJI MAGNETIC CONTACTOR SC-N2, 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (38, N'OPTIONAL', N'FUJI MAGNETIC CONTACTOR SC-N2S 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (39, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-0 (5-8A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (40, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-05 (5-8A) 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (41, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-05/3H (6-9A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (42, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-05/3H (7-11A) 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (43, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-5-1/3H (16-22A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (44, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-5-1/3H (7-11A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (45, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-5-1/3H (9-13A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (46, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-N1-3/H (24A-36A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (47, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-N1/3H (18-26A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (48, N'OPTIONAL', N'FUJI MAGNETIC STARTER SW-N2S 3/H (34-50A), 220V', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (49, N'OPTIONAL', N'FUJI MOLDED CIRCUIT BREAKER 70AT,3P, W/ NEMA 3R ENCLOSURE', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (50, N'OPTIONAL', N'FUJI T.O.R TR-0N/3 5-8A', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (51, N'OPTIONAL', N'FUJI T.O.R TR-5-1N/3 16A-22A', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (52, N'OPTIONAL', N'FUJI T.O.R TR-N3/3 (18 - 26A)', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (53, N'OPTIONAL', N'FUJI T.O.R. TR-0N/3 7-11A', N'ELEC', N'NB', N'', N'FUJ', N'Brand NB -> FUJ (housekeeping; nothing depends on it)'),
    (54, N'OPTIONAL', N'LS CIRCUIT BREAKER 10A, 2POLE, 400V', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (55, N'OPTIONAL', N'LS CIRCUIT BREAKER 20A 2P 220V ABE32B', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (56, N'OPTIONAL', N'LS CIRCUIT BREAKER ABN203, 125A 3 POLE 600V', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (57, N'OPTIONAL', N'LS MAGNETIC CONTACTOR MC40 220V', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (58, N'OPTIONAL', N'LS MAGNETIC STARTER MW-12aMB (9-13A) 220V', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (59, N'OPTIONAL', N'LS MAGNETIC STARTER MW22B W/ T.O.R MT-32/3H (16-22A) W/ ENCL. & PB', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (60, N'OPTIONAL', N'LS MAGNETIC STARTER MW40EH W/ TOR (22-32A}', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (61, N'OPTIONAL', N'LS MAGNETIC STARTER MWAM (2.5-4AMP)', N'ELEC', N'NB', N'', N'LS', N'Brand NB -> LS (housekeeping; nothing depends on it)'),
    (62, N'OPTIONAL', N'OMRON LIQUID LEVEL CONTROL 61FG (UP)', N'LLC', N'NB', N'', N'OMR', N'Brand NB -> OMR (housekeeping; nothing depends on it)'),
    (63, N'OPTIONAL', N'OMRON RELAY MY2N-GS, 220V, W/ SOCKET', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR (housekeeping; nothing depends on it)'),
    (64, N'OPTIONAL', N'OMRON RELAY MY2N-GS, 24V, DC, W/ SOCKET', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR (housekeeping; nothing depends on it)'),
    (65, N'OPTIONAL', N'OMRON RELAY MY4N-JS, 220V, W/ SOCKET', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR (housekeeping; nothing depends on it)'),
    (66, N'OPTIONAL', N'OMRON RELAY SOCKET PF-083A', N'ELEC', N'NB', N'', N'OMR', N'Brand NB -> OMR (housekeeping; nothing depends on it)'),
    (67, N'OPTIONAL', N'TELEMECANIQUE ALTERNATING RELAY 220V', N'ELEC', N'NB', N'', N'SCHN', N'Brand NB -> SCHN (housekeeping; nothing depends on it)');

BEGIN TRANSACTION;

-- Find each one exactly as the workbook saw it -----------------------------------
UPDATE a
   SET a.item_id = i.id
  FROM #action a
  JOIN tbl_setup_item i        ON i.item_model = a.item_model
  JOIN tbl_setup_item_name n   ON n.id = i.item_name_id  AND n.code = a.name_code
  JOIN tbl_setup_item_brand b  ON b.id = i.item_brand_id AND b.code = a.brand_code;

INSERT INTO #done (row_no, outcome)
SELECT row_no, N'not found as described - skipped' FROM #action WHERE item_id IS NULL;

-- 1. The edits -------------------------------------------------------------------
IF @apply_edits = 1
BEGIN
    UPDATE i
       SET i.item_brand_id = ISNULL(nb.id, i.item_brand_id),
           i.item_name_id  = ISNULL(nn.id, i.item_name_id)
      FROM tbl_setup_item i
      JOIN #action a ON a.item_id = i.id
      LEFT JOIN tbl_setup_item_brand nb ON nb.code = NULLIF(a.new_brand, N'')
      LEFT JOIN tbl_setup_item_name  nn ON nn.code = NULLIF(a.new_name,  N'')
     WHERE a.kind = N'EDIT' AND a.item_id IS NOT NULL;

    INSERT INTO #done (row_no, outcome)
    SELECT row_no, N'corrected' FROM #action
     WHERE kind = N'EDIT' AND item_id IS NOT NULL;
END
ELSE
    INSERT INTO #done (row_no, outcome)
    SELECT row_no, N'switched off' FROM #action WHERE kind = N'EDIT' AND item_id IS NOT NULL;

-- 2. The optional brand fills -----------------------------------------------------
IF @apply_optional_brands = 1
BEGIN
    UPDATE i
       SET i.item_brand_id = nb.id
      FROM tbl_setup_item i
      JOIN #action a ON a.item_id = i.id
      JOIN tbl_setup_item_brand nb ON nb.code = a.new_brand
     WHERE a.kind = N'OPTIONAL' AND a.item_id IS NOT NULL AND a.new_brand <> N'';

    INSERT INTO #done (row_no, outcome)
    SELECT row_no, N'brand set' FROM #action
     WHERE kind = N'OPTIONAL' AND item_id IS NOT NULL;
END
ELSE
    INSERT INTO #done (row_no, outcome)
    SELECT row_no, N'switched off' FROM #action WHERE kind = N'OPTIONAL' AND item_id IS NOT NULL;

-- 3. The deletions ----------------------------------------------------------------
-- Anything pointing at the item stops its deletion. The tables are found from the
-- catalogue rather than listed here, so a database with more of them is still covered:
-- every table with an item_id column, minus the item's own child rows and the audit
-- copies. based_id is deliberately not searched - on the BPI tables it means a business
-- partner, not an item.
IF @apply_deletes = 1
BEGIN
    DECLARE @tbl sysname, @sql nvarchar(max);
    DECLARE tables CURSOR LOCAL FAST_FORWARD FOR
        SELECT t.name
          FROM sys.tables t
          JOIN sys.columns c ON c.object_id = t.object_id AND c.name = 'item_id'
         WHERE t.name NOT LIKE 'z[_]%' AND t.name NOT LIKE 'zz[_]%'
           AND t.name NOT IN (N'tbl_setup_item_trade_type');

    OPEN tables;
    FETCH NEXT FROM tables INTO @tbl;
    WHILE @@FETCH_STATUS = 0
    BEGIN
        SET @sql = N'INSERT INTO #refs (item_id, referenced_by)
                     SELECT DISTINCT a.item_id, ' + QUOTENAME(@tbl, '''') + N'
                       FROM #action a
                       JOIN ' + QUOTENAME(@tbl) + N' r ON r.item_id = a.item_id
                      WHERE a.kind = N''DELETE'' AND a.item_id IS NOT NULL;';
        EXEC sp_executesql @sql;
        FETCH NEXT FROM tables INTO @tbl;
    END
    CLOSE tables;
    DEALLOCATE tables;

    INSERT INTO #done (row_no, outcome)
    SELECT a.row_no, N'still referenced - NOT deleted'
      FROM #action a
     WHERE a.kind = N'DELETE' AND a.item_id IS NOT NULL
       AND EXISTS (SELECT 1 FROM #refs r WHERE r.item_id = a.item_id);

    DECLARE @removable TABLE (item_id bigint PRIMARY KEY);
    INSERT INTO @removable (item_id)
    SELECT a.item_id FROM #action a
     WHERE a.kind = N'DELETE' AND a.item_id IS NOT NULL
       AND NOT EXISTS (SELECT 1 FROM #refs r WHERE r.item_id = a.item_id);

    DELETE FROM tbl_setup_item_additional_specs_pump_type
     WHERE additional_specs_id IN (SELECT id FROM tbl_setup_item_additional_specs
                                    WHERE based_id IN (SELECT item_id FROM @removable));
    DELETE FROM tbl_setup_item_specs_template
     WHERE based_id IN (SELECT id FROM tbl_setup_item_specs
                         WHERE based_id IN (SELECT item_id FROM @removable));
    DELETE FROM tbl_setup_item_specs            WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item_additional_specs WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item_image            WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item_inventory        WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item_model            WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_item_trade_type             WHERE based_id IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item_trade_type       WHERE item_id  IN (SELECT item_id FROM @removable);
    DELETE FROM tbl_setup_item                  WHERE id       IN (SELECT item_id FROM @removable);

    INSERT INTO #done (row_no, outcome)
    SELECT a.row_no, N'deleted' FROM #action a
     WHERE a.kind = N'DELETE' AND a.item_id IN (SELECT item_id FROM @removable);
END
ELSE
    INSERT INTO #done (row_no, outcome)
    SELECT row_no, N'switched off' FROM #action WHERE kind = N'DELETE' AND item_id IS NOT NULL;

-- What happened -------------------------------------------------------------------
SELECT a.kind, d.outcome, COUNT(*) AS rows_
  FROM #action a JOIN #done d ON d.row_no = a.row_no
 GROUP BY a.kind, d.outcome ORDER BY a.kind, d.outcome;

SELECT a.kind, d.outcome, a.item_model, a.name_code AS was_name, a.brand_code AS was_brand,
       NULLIF(a.new_name, N'') AS now_name, NULLIF(a.new_brand, N'') AS now_brand, a.why
  FROM #action a JOIN #done d ON d.row_no = a.row_no
 ORDER BY a.kind, a.item_model;

IF EXISTS (SELECT 1 FROM #refs)
    SELECT DISTINCT 'blocked the delete' AS note, a.item_model, r.referenced_by
      FROM #refs r JOIN #action a ON a.item_id = r.item_id
     ORDER BY a.item_model, r.referenced_by;

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

IF OBJECT_ID('tempdb..#action') IS NOT NULL DROP TABLE #action;
IF OBJECT_ID('tempdb..#done')   IS NOT NULL DROP TABLE #done;
IF OBJECT_ID('tempdb..#refs')   IS NOT NULL DROP TABLE #refs;
