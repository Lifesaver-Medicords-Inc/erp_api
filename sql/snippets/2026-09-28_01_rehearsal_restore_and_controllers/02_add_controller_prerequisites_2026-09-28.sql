-- STEP 02 of 04 - what the controller batch needs on a database that was not built from test_fresh
-- Run the numbered steps in order, each on its own:
--   01  01_restore_rehearsal_setup_from_before_wipe_2026-09-28.sql (rehearsal only - skip elsewhere)
--   02  this file
--   03  ..\2026-09-25_01_controllers_and_project_quote\05_load_controllers_items_2026-09-23.sql   (@commit = 1)
--   04  ..\2026-09-25_01_controllers_and_project_quote\07_load_controllers_bom_2026-09-23.sql     (@commit = 1, @round_up_fractional_qty = 1)
-- Tested 2026-09-28 end to end (02, 03, 04 committed) on a point-in-time copy of rehearsal:
-- 5,975 items, 5,840 BOMs, 130,372 lines (4,096 wire lengths rounded up), 39,180 remarks.
-- 06_fix_qerp_items_for_controllers is NOT needed after this file: the 17 catalogue items below
-- are added already corrected, so 06 would only report "not on this database" for every row.
-- Every step is a TRIAL RUN until you set @commit = 1; it must end by printing "committed".

-- Controller prerequisites for Lightspeed_ERP_deploy_rehearsal (and any database like it)
-- ====================================================================================
-- The controller batch (2026-09-25_01, steps 05-07) was written against test_fresh's Setup.
-- On rehearsal 05 stops on 31 missing Setup codes and 07 on 17 missing catalogue items
-- (checked 2026-09-28 on a point-in-time copy of rehearsal). This adds exactly those, with
-- the codes and names test_fresh uses:
--
--   Item Brand   'No Brand' is NBR on rehearsal and NB on test_fresh. The code is RENAMED to
--                NB (same row, same id, same name) rather than a second 'No Brand' being
--                added - nothing in the code or the scripts refers to 'NBR'. Then 15 brands:
--                AND DAL FUJ JKN LS PHF SCHN SKD SQD YKY (05) and ANLY CHS MAC OMR WIX (07).
--   Item Name    AR ELEC MC MCB MCCB OUV PB PL SS TB TL TOR (05) and HDW ME (07).
--   Pump Type    BP DP TP.            Pump Count   SMP DPX TPX QPX QPX2.
--   Items        the 17 QERP items the controller BOMs use - 9 control boxes, 3 THHN wires,
--                the ANLY timer, the Omron relay, the Wixim tie mount, the Mac 3 float switch
--                and the CHS cable tie - as they stand on test_fresh after its 06 fixes.
--                Each gets its trade type TRD and its long description, as the QERP load
--                wrote them. item_tangibility_type is TANGIBLE (test_fresh carries 'TRD' there
--                on 1,355 QERP rows, which is not a tangibility). item_code continues from the
--                highest issued, as the API does.
--
-- Additive and safe to run again: a code or item that is already there is skipped and
-- reported; nothing else is updated or deleted.
--
-- HOW TO RUN: sqlcmd -S <server> -d <database> -b -I -i <this file>  (or open it in SSMS,
-- and CHECK THE DATABASE DROPDOWN FIRST). With @commit = 0 everything is done and reported,
-- then rolled back. After a commit, restart that database's API or clear its Redis keys.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

PRINT N'Database: ' + DB_NAME();

DECLARE @done TABLE (what nvarchar(40), code nvarchar(400), outcome nvarchar(80));

-- Lookups the items need that this file does not add. Stop before writing if one is missing.
DECLARE @missing TABLE (setup_list nvarchar(40), code nvarchar(50));
INSERT INTO @missing
SELECT v.list, v.code
  FROM (VALUES (N'Item Class', N'HW'), (N'Unit of Measure', N'PC'), (N'Unit of Measure', N'PACK'),
               (N'Unit of Measure', N'MTR'), (N'Item Type', N'TRD'), (N'Item Name', N'FS'),
               (N'Item Name', N'WMAT')) v(list, code)
 WHERE (v.list = N'Item Class'      AND NOT EXISTS (SELECT 1 FROM tbl_setup_item_class x            WHERE x.code = v.code))
    OR (v.list = N'Unit of Measure' AND NOT EXISTS (SELECT 1 FROM tbl_setup_item_unit_measurement x WHERE x.code = v.code))
    OR (v.list = N'Item Type'       AND NOT EXISTS (SELECT 1 FROM tbl_setup_item_type x             WHERE x.code = v.code))
    OR (v.list = N'Item Name'       AND NOT EXISTS (SELECT 1 FROM tbl_setup_item_name x             WHERE x.code = v.code));

IF EXISTS (SELECT 1 FROM @missing)
BEGIN
    SELECT N'missing from Setup - nothing was written' AS problem, setup_list, code FROM @missing;
    RAISERROR(N'Setup rows are missing (listed above). Add them, then run this again.', 16, 1);
    RETURN;
END

BEGIN TRANSACTION;

-- 1. 'No Brand': one row, test_fresh's code ------------------------------------------------
IF EXISTS (SELECT 1 FROM tbl_setup_item_brand WHERE code = N'NB')
    INSERT INTO @done VALUES (N'Item Brand', N'NB', N'already there');
ELSE IF EXISTS (SELECT 1 FROM tbl_setup_item_brand WHERE code = N'NBR')
BEGIN
    UPDATE tbl_setup_item_brand SET code = N'NB' WHERE code = N'NBR';
    INSERT INTO @done VALUES (N'Item Brand', N'NBR -> NB', N'code renamed (same row, same id)');
END
ELSE
BEGIN
    INSERT INTO tbl_setup_item_brand (code, name) VALUES (N'NB', N'No Brand');
    INSERT INTO @done VALUES (N'Item Brand', N'NB', N'added');
END

-- 2. Brands ---------------------------------------------------------------------------------
DECLARE @brand TABLE (code nvarchar(50), name nvarchar(200));
INSERT INTO @brand VALUES
    (N'AND', N'ANDELI'), (N'DAL', N'Dalier'), (N'FUJ', N'Fuji'), (N'JKN', N'JKN Under'), (N'LS', N'LS'),
    (N'PHF', N'Philflex'), (N'SCHN', N'Schneider'), (N'SKD', N'Sakura Denso'), (N'SQD', N'Square D'),
    (N'YKY', N'Yokoyama'), (N'ANLY', N'ANLY'), (N'CHS', N'CHS'), (N'MAC', N'Mac 3'), (N'OMR', N'Omron'),
    (N'WIX', N'Wixim');

INSERT INTO @done SELECT N'Item Brand', b.code, N'already there'
  FROM @brand b WHERE EXISTS (SELECT 1 FROM tbl_setup_item_brand x WHERE x.code = b.code);
INSERT INTO tbl_setup_item_brand (code, name)
SELECT b.code, b.name FROM @brand b WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_item_brand x WHERE x.code = b.code);
INSERT INTO @done SELECT N'Item Brand', b.code, N'added'
  FROM @brand b WHERE NOT EXISTS (SELECT 1 FROM @done d WHERE d.what = N'Item Brand' AND d.code = b.code);

-- 3. Item names -----------------------------------------------------------------------------
DECLARE @name TABLE (code nvarchar(50), name nvarchar(200));
INSERT INTO @name VALUES
    (N'AR', N'ALTERNATING RELAY'), (N'ELEC', N'ELECTRICAL'), (N'MC', N'MAGNETIC CONTACTOR'),
    (N'MCB', N'MINIATURE CIRCUIT BREAKER'), (N'MCCB', N'MOLDED CASE CIRCUIT BREAKER'), (N'OUV', N'O/U VOLTAGE'),
    (N'PB', N'PUSH BUTTON'), (N'PL', N'PILOT LIGHT'), (N'SS', N'SELECTOR SWITCH'), (N'TB', N'TERMINAL BLOCK'),
    (N'TL', N'TERMINAL LUG'), (N'TOR', N'THERMAL OVERLOAD RELAY'), (N'HDW', N'HARDWARE'), (N'ME', N'METAL ENCLOSURE');

INSERT INTO @done SELECT N'Item Name', n.code, N'already there'
  FROM @name n WHERE EXISTS (SELECT 1 FROM tbl_setup_item_name x WHERE x.code = n.code);
INSERT INTO tbl_setup_item_name (code, name)
SELECT n.code, n.name FROM @name n WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_item_name x WHERE x.code = n.code);
INSERT INTO @done SELECT N'Item Name', n.code, N'added'
  FROM @name n WHERE NOT EXISTS (SELECT 1 FROM @done d WHERE d.what = N'Item Name' AND d.code = n.code);

-- 4. Pump types and pump counts ---------------------------------------------------------------
DECLARE @ptype TABLE (code nvarchar(50), name nvarchar(200));
INSERT INTO @ptype VALUES (N'BP', N'BOOSTER PUMP'), (N'DP', N'DRAINAGE PUMP'), (N'TP', N'TRANSFER PUMP');

INSERT INTO @done SELECT N'Pump Type', p.code, N'already there'
  FROM @ptype p WHERE EXISTS (SELECT 1 FROM tbl_setup_item_pump_type x WHERE x.code = p.code);
INSERT INTO tbl_setup_item_pump_type (code, name)
SELECT p.code, p.name FROM @ptype p WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_item_pump_type x WHERE x.code = p.code);
INSERT INTO @done SELECT N'Pump Type', p.code, N'added'
  FROM @ptype p WHERE NOT EXISTS (SELECT 1 FROM @done d WHERE d.what = N'Pump Type' AND d.code = p.code);

DECLARE @pcount TABLE (code nvarchar(50), name nvarchar(200));
INSERT INTO @pcount VALUES (N'SMP', N'Simplex'), (N'DPX', N'Duplex'), (N'TPX', N'Triplex'), (N'QPX', N'Quadruplex'),
                           (N'QPX2', N'Quintuplex');

INSERT INTO @done SELECT N'Pump Count', p.code, N'already there'
  FROM @pcount p WHERE EXISTS (SELECT 1 FROM tbl_setup_item_pump_count x WHERE x.code = p.code);
INSERT INTO tbl_setup_item_pump_count (code, name)
SELECT p.code, p.name FROM @pcount p WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_item_pump_count x WHERE x.code = p.code);
INSERT INTO @done SELECT N'Pump Count', p.code, N'added'
  FROM @pcount p WHERE NOT EXISTS (SELECT 1 FROM @done d WHERE d.what = N'Pump Count' AND d.code = p.code);

-- 5. The 17 catalogue items the controller BOMs use ---------------------------------------------
DECLARE @item TABLE (row_no int PRIMARY KEY, name_code nvarchar(50), brand_code nvarchar(50), class_code nvarchar(50),
                     uom_code nvarchar(50), item_model nvarchar(400), item_id bigint NULL);
INSERT INTO @item (row_no, name_code, brand_code, class_code, uom_code, item_model) VALUES
    (1,  N'ELEC', N'ANLY', N'HW', N'PC',   N'ANLY STAR DELTA TIMER TRD-N 220V W/ SOCKET'),
    (2,  N'ELEC', N'OMR',  N'HW', N'PC',   N'OMRON FLOATLESS RELAY 61FG (DOWN)'),
    (3,  N'ELEC', N'WIX',  N'HW', N'PACK', N'TIE MOUNT 1" X 1"'),
    (4,  N'FS',   N'MAC',  N'HW', N'PC',   N'MAC 3 FLOAT SWITCH'),
    (5,  N'HDW',  N'CHS',  N'HW', N'PACK', N'CABLE TIE 4"'),
    (6,  N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 16" X 14" X 8" N3R'),
    (7,  N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 20" X 17" X 9" N3R'),
    (8,  N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 20" X 20" X 9" N3R'),
    (9,  N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 25" X 20" X 9" NR3'),
    (10, N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 28" X 22" X 9" N3R'),
    (11, N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 28" X 24" X 10" N3R'),
    (12, N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 36" X 26" X 14" N3R W/ VIEWING GLASS, WRINKLED BEIGE'),
    (13, N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 400MM X 300MM X 200MM'),
    (14, N'ME',   N'NB',   N'HW', N'PC',   N'CONTROL BOX 600MM X 500MM X 200MM'),
    (15, N'WMAT', N'NB',   N'HW', N'MTR',  N'THHN WIRE 2.0MMSQ (MTR)'),
    (16, N'WMAT', N'NB',   N'HW', N'MTR',  N'THHN WIRE 5.5MMSQ. STRANDED (MTR)'),
    (17, N'WMAT', N'NB',   N'HW', N'MTR',  N'THHN WIRE 8.00MM SQ. STRANDED (MTR)');

-- Already there = same item name + brand + model, the key 07 matches components on.
UPDATE i SET item_id = x.id
  FROM @item i
  JOIN tbl_setup_item_name n  ON n.code = i.name_code
  JOIN tbl_setup_item_brand b ON b.code = i.brand_code
  JOIN tbl_setup_item x       ON x.item_name_id = n.id AND x.item_brand_id = b.id AND x.item_model = i.item_model;

INSERT INTO @done SELECT N'Item', item_model, N'already there' FROM @item WHERE item_id IS NOT NULL;

-- A model already used under a different brand would make a second catalogue row with the
-- same model; refuse rather than guess (spec 4.2.2 - models repeat only across brands on purpose).
IF EXISTS (SELECT 1 FROM @item i JOIN tbl_setup_item x ON x.item_model = i.item_model WHERE i.item_id IS NULL)
BEGIN
    SELECT N'model already in the catalogue under another name/brand - nothing was written' AS problem,
           i.item_model, n.code AS name_code_there, b.code AS brand_code_there
      FROM @item i
      JOIN tbl_setup_item x ON x.item_model = i.item_model
      LEFT JOIN tbl_setup_item_name n ON n.id = x.item_name_id
      LEFT JOIN tbl_setup_item_brand b ON b.id = x.item_brand_id
     WHERE i.item_id IS NULL;
    ROLLBACK TRANSACTION;
    RAISERROR(N'Catalogue conflict (listed above) - rolled back, nothing changed.', 16, 1);
    RETURN;
END

DECLARE @trade_id bigint = (SELECT id FROM tbl_setup_item_type WHERE code = N'TRD');
DECLARE @next int = ISNULL((SELECT MAX(TRY_CAST(item_code AS int)) FROM tbl_setup_item), 0) + 1;
DECLARE @row int = (SELECT MIN(row_no) FROM @item WHERE item_id IS NULL), @id bigint;

WHILE @row IS NOT NULL
BEGIN
    INSERT INTO tbl_setup_item (item_name_id, item_class_id, item_brand_id, unit_of_measure_id, item_model,
                                catalogue_year, item_code, item_tangibility_type, is_stop_selling, price)
    SELECT n.id, c.id, b.id, u.id, i.item_model, NULL, RIGHT(N'0000' + CAST(@next AS nvarchar(20)), 4),
           N'TANGIBLE', 0, 0
      FROM @item i
      JOIN tbl_setup_item_name n             ON n.code = i.name_code
      JOIN tbl_setup_item_brand b            ON b.code = i.brand_code
      JOIN tbl_setup_item_class c            ON c.code = i.class_code
      JOIN tbl_setup_item_unit_measurement u ON u.code = i.uom_code
     WHERE i.row_no = @row;

    SET @id = SCOPE_IDENTITY();
    UPDATE @item SET item_id = @id WHERE row_no = @row;

    INSERT INTO tbl_setup_item_trade_type (item_id, trade_type_id) VALUES (@id, @trade_id);
    INSERT INTO tbl_setup_item_additional_specs (based_id, long_description)
    SELECT @id, item_model FROM @item WHERE row_no = @row;

    INSERT INTO @done SELECT N'Item', item_model, N'added, item code ' + RIGHT(N'0000' + CAST(@next AS nvarchar(20)), 4)
      FROM @item WHERE row_no = @row;

    SET @next = @next + 1;
    SET @row = (SELECT MIN(row_no) FROM @item WHERE item_id IS NULL);
END

SELECT what, code, outcome FROM @done ORDER BY what, outcome, code;

IF EXISTS (SELECT 1 FROM @item WHERE item_id IS NULL)
BEGIN
    ROLLBACK TRANSACTION;
    RAISERROR(N'An item could not be written - rolled back, nothing changed.', 16, 1);
    RETURN;
END

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT N'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT N'trial run - rolled back, nothing changed. Set @commit = 1 to keep it.';
END
