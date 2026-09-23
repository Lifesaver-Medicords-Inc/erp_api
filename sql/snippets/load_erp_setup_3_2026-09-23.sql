-- Setup lists from "ERP SETUP_3.xlsx" (data migration excel format)
-- =====================================================================
-- Adds the entries that are not in the setup tables yet and leaves everything else
-- alone: a row counts as already present when its CODE or its NAME matches, case and
-- surrounding spaces ignored, so a list someone half-entered by hand is topped up
-- rather than duplicated. Nothing is updated or deleted - an existing row keeps the
-- code and name it has, and the script reports where the file disagrees with it.
--
-- id is an identity column on every table here, so only (code, name) is written.
--
-- NOTE on the Position sheet: this system keeps ONE position list. tbl_position holds
-- the employee positions that drive Access Control and the per-position navigation, and
-- tbl_bpi_contacts.position points at that same table for a contact's role. The sheet's
-- entries (Owner, Contractor, Project Manager, ...) are contact roles, so they will also
-- show up in Admin's Access Control as positions that can be granted access. They are
-- inserted because that is how the two screens already share the table; the script lists
-- them at the end so it stays a visible decision rather than a side effect.
--
-- Set @commit = 1 to write. With @commit = 0 it reports and rolls back.

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;

IF OBJECT_ID('tempdb..#incoming') IS NOT NULL DROP TABLE #incoming;
IF OBJECT_ID('tempdb..#added') IS NOT NULL DROP TABLE #added;
CREATE TABLE #incoming (table_name sysname, code nvarchar(200), name nvarchar(400));
CREATE TABLE #added    (table_name sysname, code nvarchar(200), name nvarchar(400));

BEGIN TRANSACTION;

-- Everything the file carries -------------------------------------------------

-- Item Brand (27)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_brand', N'ACE', N'Acejet'),
    (N'tbl_setup_item_brand', N'AND', N'ANDELI'),
    (N'tbl_setup_item_brand', N'ANLY', N'ANLY'),
    (N'tbl_setup_item_brand', N'CLP', N'Calpeda'),
    (N'tbl_setup_item_brand', N'CHS', N'CHS'),
    (N'tbl_setup_item_brand', N'DAL', N'Dalier'),
    (N'tbl_setup_item_brand', N'DAN', N'Danfoss'),
    (N'tbl_setup_item_brand', N'DEL', N'Delta'),
    (N'tbl_setup_item_brand', N'DYNA', N'Dynaflo'),
    (N'tbl_setup_item_brand', N'ETN', N'EATON'),
    (N'tbl_setup_item_brand', N'FTI', N'Finish Thompson Inc.'),
    (N'tbl_setup_item_brand', N'FUJ', N'Fuji'),
    (N'tbl_setup_item_brand', N'ISP', N'Instapure'),
    (N'tbl_setup_item_brand', N'JKN', N'JKN Under'),
    (N'tbl_setup_item_brand', N'LGI', N'Little Giant'),
    (N'tbl_setup_item_brand', N'LS', N'LS'),
    (N'tbl_setup_item_brand', N'MAC', N'Mac 3'),
    (N'tbl_setup_item_brand', N'NB', N'No Brand'),
    (N'tbl_setup_item_brand', N'OMR', N'Omron'),
    (N'tbl_setup_item_brand', N'PTA', N'Pentair Aurora'),
    (N'tbl_setup_item_brand', N'PHF', N'Philflex'),
    (N'tbl_setup_item_brand', N'SKD', N'Sakura Denso'),
    (N'tbl_setup_item_brand', N'SCHN', N'Schneider'),
    (N'tbl_setup_item_brand', N'SHN', N'Shinmaywa'),
    (N'tbl_setup_item_brand', N'SQD', N'Square D'),
    (N'tbl_setup_item_brand', N'WIX', N'Wixim'),
    (N'tbl_setup_item_brand', N'YKY', N'Yokoyama');

-- Item Class (12)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_class', N'TAI', N'TOOLS AND INSTRUMENTS'),
    (N'tbl_setup_item_class', N'CON', N'CONSTRUCTION'),
    (N'tbl_setup_item_class', N'ELEC', N'ELECTRICAL'),
    (N'tbl_setup_item_class', N'EXP', N'EXPENSE (EMPTY SPECS)'),
    (N'tbl_setup_item_class', N'HW', N'HARDWARE'),
    (N'tbl_setup_item_class', N'IMP', N'IMPORT PURCHASE'),
    (N'tbl_setup_item_class', N'LOC', N'LOCAL PURCHASE'),
    (N'tbl_setup_item_class', N'OFF', N'OFFICE'),
    (N'tbl_setup_item_class', N'PMB', N'PLUMBING'),
    (N'tbl_setup_item_class', N'SER', N'SERVICES'),
    (N'tbl_setup_item_class', N'WM', N'WATER METER'),
    (N'tbl_setup_item_class', N'PMPP', N'PUMP PARTS');

-- Item Material (10)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_material', N'AISL 304', N'STAINLESS STEEL 304'),
    (N'tbl_setup_item_material', N'AISL 316', N'STAINLESS STEEL 316'),
    (N'tbl_setup_item_material', N'DI', N'DUCTILE IRON'),
    (N'tbl_setup_item_material', N'BI', N'BLACK IRON'),
    (N'tbl_setup_item_material', N'GB', N'BRONZE'),
    (N'tbl_setup_item_material', N'GG', N'CAST IRON'),
    (N'tbl_setup_item_material', N'GI', N'GALVANIZED IRON'),
    (N'tbl_setup_item_material', N'GO', N'BRASS'),
    (N'tbl_setup_item_material', N'PPO-GF20', N'NORYL'),
    (N'tbl_setup_item_material', N'FS', N'FORGED STEEL');

-- Item Name (47)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_name', N'ARV', N'AIR RELEASE VALVE'),
    (N'tbl_setup_item_name', N'AR', N'ALTERNATING RELAY'),
    (N'tbl_setup_item_name', N'BV', N'BALL VALVE'),
    (N'tbl_setup_item_name', N'BT', N'BLADDER TANK'),
    (N'tbl_setup_item_name', N'BFV', N'BUTTERFLY VALVE'),
    (N'tbl_setup_item_name', N'CP', N'COMMON PACKAGE'),
    (N'tbl_setup_item_name', N'CONS', N'CONSUMABLES'),
    (N'tbl_setup_item_name', N'CTRL', N'CONTROLLER'),
    (N'tbl_setup_item_name', N'DCH', N'DISCHARGE COMMON HEADER'),
    (N'tbl_setup_item_name', N'DRB', N'DISCHARGE RUBBER BELLOW'),
    (N'tbl_setup_item_name', N'ELEC', N'ELECTRICAL'),
    (N'tbl_setup_item_name', N'FS', N'FLOAT SWITCH'),
    (N'tbl_setup_item_name', N'FLV', N'FLOAT VALVE'),
    (N'tbl_setup_item_name', N'FM', N'FLOW METER'),
    (N'tbl_setup_item_name', N'FTV', N'FOOT VALVE'),
    (N'tbl_setup_item_name', N'GV', N'GATE VALVE'),
    (N'tbl_setup_item_name', N'HDW', N'HARDWARE'),
    (N'tbl_setup_item_name', N'LLC', N'LIQUID LEVEL CONTROL'),
    (N'tbl_setup_item_name', N'MC', N'MAGNETIC CONTACTOR'),
    (N'tbl_setup_item_name', N'ME', N'METAL ENCLOSURE'),
    (N'tbl_setup_item_name', N'MCB', N'MINIATURE CIRCUIT BREAKER'),
    (N'tbl_setup_item_name', N'MCCB', N'MOLDED CASE CIRCUIT BREAKER'),
    (N'tbl_setup_item_name', N'OUV', N'O/U VOLTAGE'),
    (N'tbl_setup_item_name', N'OTH', N'OTHER'),
    (N'tbl_setup_item_name', N'PNT', N'PAINT'),
    (N'tbl_setup_item_name', N'PL', N'PILOT LIGHT'),
    (N'tbl_setup_item_name', N'PG', N'PRESSURE GAUGE'),
    (N'tbl_setup_item_name', N'PRV', N'PRESSURE RELIEF VALVE'),
    (N'tbl_setup_item_name', N'PS', N'PRESSURE SWITCH'),
    (N'tbl_setup_item_name', N'PT', N'PRESSURE TRANSDUCER'),
    (N'tbl_setup_item_name', N'PMP', N'PUMP'),
    (N'tbl_setup_item_name', N'PMPP', N'PUMP PART'),
    (N'tbl_setup_item_name', N'PB', N'PUSH BUTTON'),
    (N'tbl_setup_item_name', N'SS', N'SELECTOR SWITCH'),
    (N'tbl_setup_item_name', N'SER', N'SERVICE'),
    (N'tbl_setup_item_name', N'SCV', N'SUCTION CHECK VALVE'),
    (N'tbl_setup_item_name', N'SCH', N'SUCTION COMMON HEADER'),
    (N'tbl_setup_item_name', N'SRB', N'SUCTION RUBBER BELLOW'),
    (N'tbl_setup_item_name', N'SUP', N'SUPPLIES'),
    (N'tbl_setup_item_name', N'TB', N'TERMINAL BLOCK'),
    (N'tbl_setup_item_name', N'TL', N'TERMINAL LUG'),
    (N'tbl_setup_item_name', N'TOR', N'THERMAL OVERLOAD RELAY'),
    (N'tbl_setup_item_name', N'TAI', N'TOOLS AND INSTRUMENTS'),
    (N'tbl_setup_item_name', N'VFD', N'VARIABLE FREQUENCY DRIVE'),
    (N'tbl_setup_item_name', N'WC', N'WASTE CONE'),
    (N'tbl_setup_item_name', N'WM', N'WATER METER'),
    (N'tbl_setup_item_name', N'WMAT', N'WIRING MATERIALS');

-- Item Pump Count (5)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_pump_count', N'SMP', N'Simplex'),
    (N'tbl_setup_item_pump_count', N'DPX', N'Duplex'),
    (N'tbl_setup_item_pump_count', N'TPX', N'Triplex'),
    (N'tbl_setup_item_pump_count', N'QPX', N'Quadruplex'),
    (N'tbl_setup_item_pump_count', N'QPX2', N'Quintuplex');

-- Item Pump Type (7)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_pump_type', N'BP', N'BOOSTER PUMP'),
    (N'tbl_setup_item_pump_type', N'CPS', N'CPS'),
    (N'tbl_setup_item_pump_type', N'DP', N'DRAINAGE PUMP'),
    (N'tbl_setup_item_pump_type', N'NUFP', N'NON-ULFM FIRE PUMP'),
    (N'tbl_setup_item_pump_type', N'NUJP', N'NON-ULFM JOCKEY PUMP'),
    (N'tbl_setup_item_pump_type', N'SP', N'SEWAGE PUMP'),
    (N'tbl_setup_item_pump_type', N'TP', N'TRANSFER PUMP');

-- Item Type (2)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_type', N'TRD', N'TRADE'),
    (N'tbl_setup_item_type', N'NTR', N'NON-TRADE');

-- Unit of Measure (29)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_item_unit_measurement', N'BAG', N'Bag'),
    (N'tbl_setup_item_unit_measurement', N'BKS', N'Books'),
    (N'tbl_setup_item_unit_measurement', N'BTL', N'Bottle'),
    (N'tbl_setup_item_unit_measurement', N'CAN', N'Can'),
    (N'tbl_setup_item_unit_measurement', N'CBY', N'Cubic Yard'),
    (N'tbl_setup_item_unit_measurement', N'DOZ.', N'Dozen'),
    (N'tbl_setup_item_unit_measurement', N'DRM', N'Drum'),
    (N'tbl_setup_item_unit_measurement', N'FEET', N'Feet'),
    (N'tbl_setup_item_unit_measurement', N'GAL', N'Gallon'),
    (N'tbl_setup_item_unit_measurement', N'GR.', N'Gross'),
    (N'tbl_setup_item_unit_measurement', N'KEG', N'Keg'),
    (N'tbl_setup_item_unit_measurement', N'KG', N'Kilogram'),
    (N'tbl_setup_item_unit_measurement', N'LM', N'Linear Meter'),
    (N'tbl_setup_item_unit_measurement', N'LOT', N'Lot'),
    (N'tbl_setup_item_unit_measurement', N'LTR', N'Liter'),
    (N'tbl_setup_item_unit_measurement', N'MTR', N'Meter'),
    (N'tbl_setup_item_unit_measurement', N'PACK', N'Pack'),
    (N'tbl_setup_item_unit_measurement', N'PAD', N'Pad'),
    (N'tbl_setup_item_unit_measurement', N'PAIL', N'Pail'),
    (N'tbl_setup_item_unit_measurement', N'PC', N'Piece'),
    (N'tbl_setup_item_unit_measurement', N'PRS', N'Pairs'),
    (N'tbl_setup_item_unit_measurement', N'QAT', N'Quart'),
    (N'tbl_setup_item_unit_measurement', N'REAM', N'Ream'),
    (N'tbl_setup_item_unit_measurement', N'ROLL', N'Roll'),
    (N'tbl_setup_item_unit_measurement', N'SET', N'Set'),
    (N'tbl_setup_item_unit_measurement', N'SHEET', N'Sheet'),
    (N'tbl_setup_item_unit_measurement', N'TANK', N'Tank'),
    (N'tbl_setup_item_unit_measurement', N'TIN', N'Tin'),
    (N'tbl_setup_item_unit_measurement', N'UNIT', N'Unit');

-- Payment terms (13)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_payment_terms', N'DP50', N'Downpayment 50%'),
    (N'tbl_setup_payment_terms', N'DP25', N'Downpayment 25%'),
    (N'tbl_setup_payment_terms', N'DP75', N'Downpayment 75%'),
    (N'tbl_setup_payment_terms', N'COD', N'Cash on Delivery'),
    (N'tbl_setup_payment_terms', N'COD-C', N'Cash on Delivery - Cheque'),
    (N'tbl_setup_payment_terms', N'NET15', N'Net 15'),
    (N'tbl_setup_payment_terms', N'NET30', N'Net 30'),
    (N'tbl_setup_payment_terms', N'NET60', N'Net 60'),
    (N'tbl_setup_payment_terms', N'PDC30', N'Post Dated Cheque 30'),
    (N'tbl_setup_payment_terms', N'PDC45', N'Post Dated Cheque 45'),
    (N'tbl_setup_payment_terms', N'PDC60', N'Post Dated Cheque 60'),
    (N'tbl_setup_payment_terms', N'50-50', N'50%-50%'),
    (N'tbl_setup_payment_terms', N'30-20-50', N'30%- 20%- 50%');

-- Socials (12)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_social', N'EML', N'Email'),
    (N'tbl_setup_social', N'FB', N'Facebook'),
    (N'tbl_setup_social', N'IG', N'Instagram'),
    (N'tbl_setup_social', N'INT', N'Internet Search'),
    (N'tbl_setup_social', N'LNK', N'LinkedIn'),
    (N'tbl_setup_social', N'REF', N'Referral'),
    (N'tbl_setup_social', N'ST', N'Sales Team'),
    (N'tbl_setup_social', N'TG', N'Telegram'),
    (N'tbl_setup_social', N'TK', N'TikTok'),
    (N'tbl_setup_social', N'VBR', N'Viber'),
    (N'tbl_setup_social', N'WEB', N'Website'),
    (N'tbl_setup_social', N'WSP', N'WhatsApp');

-- Entity Type (7)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_bpi_entity', N'AFF', N'Affiliated'),
    (N'tbl_setup_bpi_entity', N'BLK', N'Blacklisted'),
    (N'tbl_setup_bpi_entity', N'CLD', N'Closed'),
    (N'tbl_setup_bpi_entity', N'CUS', N'Customer'),
    (N'tbl_setup_bpi_entity', N'NAF', N'Non-Affiliated'),
    (N'tbl_setup_bpi_entity', N'SUP', N'Supplier'),
    (N'tbl_setup_bpi_entity', N'TSP', N'Temporary Supplier');

-- Industries (26)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_bpi_industries', N'AGRI', N'Agriculture/ Irrigation'),
    (N'tbl_setup_bpi_industries', N'AQUA', N'Aquaculture'),
    (N'tbl_setup_bpi_industries', N'CHEM', N'Chemical'),
    (N'tbl_setup_bpi_industries', N'COLD', N'Cold Storage'),
    (N'tbl_setup_bpi_industries', N'CON', N'Construction'),
    (N'tbl_setup_bpi_industries', N'DATA', N'Data Centers'),
    (N'tbl_setup_bpi_industries', N'DOM', N'Domestic/ Residential'),
    (N'tbl_setup_bpi_industries', N'FILTER', N'Filtration'),
    (N'tbl_setup_bpi_industries', N'FIRE', N'Fire Systems'),
    (N'tbl_setup_bpi_industries', N'FBV', N'Food / Beverage'),
    (N'tbl_setup_bpi_industries', N'GRND', N'Groundwater'),
    (N'tbl_setup_bpi_industries', N'HC', N'Healthcare'),
    (N'tbl_setup_bpi_industries', N'HVAC', N'HVAC'),
    (N'tbl_setup_bpi_industries', N'IND', N'Industrial'),
    (N'tbl_setup_bpi_industries', N'ICE', N'Ice Plant'),
    (N'tbl_setup_bpi_industries', N'LAND', N'Landscaping'),
    (N'tbl_setup_bpi_industries', N'MRN', N'Marine'),
    (N'tbl_setup_bpi_industries', N'MINE', N'Mining'),
    (N'tbl_setup_bpi_industries', N'MUNI', N'Municipal'),
    (N'tbl_setup_bpi_industries', N'OIL', N'Oil / Gas'),
    (N'tbl_setup_bpi_industries', N'PHRM', N'Pharmaceuticals'),
    (N'tbl_setup_bpi_industries', N'PWR', N'Power Generation'),
    (N'tbl_setup_bpi_industries', N'POOL', N'Pools'),
    (N'tbl_setup_bpi_industries', N'SC', N'Semiconductors'),
    (N'tbl_setup_bpi_industries', N'SKCM', N'Skyscrapers/ Commercial'),
    (N'tbl_setup_bpi_industries', N'WT', N'Water Treatment');

-- Position (7)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_position', N'CTR', N'Contractor'),
    (N'tbl_position', N'EE', N'Electrical Engr.'),
    (N'tbl_position', N'ME', N'Mechanical Engr.'),
    (N'tbl_position', N'OWN', N'Owner'),
    (N'tbl_position', N'PM', N'Project Manager'),
    (N'tbl_position', N'PUR', N'Purchasing'),
    (N'tbl_position', N'SRP', N'Sales Representative');

-- Warehouse Use Type (5)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_inv_warehouse_usetype', N'DMG', N'DAMAGED ITEMS'),
    (N'tbl_inv_warehouse_usetype', N'GEN', N'GENERAL'),
    (N'tbl_inv_warehouse_usetype', N'OUT', N'OUTBOUND'),
    (N'tbl_inv_warehouse_usetype', N'REC', N'RECEIVING'),
    (N'tbl_inv_warehouse_usetype', N'SP', N'SPARE PARTS');

-- Valuation Method (6)
INSERT INTO #incoming (table_name, code, name) VALUES
    (N'tbl_setup_valuation_method', N'FIFO', N'First In First Out'),
    (N'tbl_setup_valuation_method', N'LIFO', N'Last In First Out'),
    (N'tbl_setup_valuation_method', N'WAC', N'Weighted Average Cost'),
    (N'tbl_setup_valuation_method', N'MAC', N'Moving Average Cost'),
    (N'tbl_setup_valuation_method', N'JIT', N'Just In Time'),
    (N'tbl_setup_valuation_method', N'STD', N'Standard Costing');

-- Insert what is missing, table by table ---------------------------------------

INSERT INTO tbl_setup_item_brand (code, name)
OUTPUT N'tbl_setup_item_brand', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_brand'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_brand t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_class (code, name)
OUTPUT N'tbl_setup_item_class', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_class'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_class t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_material (code, name)
OUTPUT N'tbl_setup_item_material', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_material'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_material t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_name (code, name)
OUTPUT N'tbl_setup_item_name', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_name'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_name t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_pump_count (code, name)
OUTPUT N'tbl_setup_item_pump_count', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_pump_count'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_pump_count t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_pump_type (code, name)
OUTPUT N'tbl_setup_item_pump_type', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_pump_type'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_pump_type t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_type (code, name)
OUTPUT N'tbl_setup_item_type', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_type'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_type t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_item_unit_measurement (code, name)
OUTPUT N'tbl_setup_item_unit_measurement', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_item_unit_measurement'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_item_unit_measurement t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_payment_terms (code, name)
OUTPUT N'tbl_setup_payment_terms', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_payment_terms'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_payment_terms t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_social (code, name)
OUTPUT N'tbl_setup_social', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_social'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_social t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_bpi_entity (code, name)
OUTPUT N'tbl_setup_bpi_entity', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_bpi_entity'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_bpi_entity t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_bpi_industries (code, name)
OUTPUT N'tbl_setup_bpi_industries', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_bpi_industries'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_bpi_industries t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_position (code, name)
OUTPUT N'tbl_position', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_position'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_position t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_inv_warehouse_usetype (code, name)
OUTPUT N'tbl_inv_warehouse_usetype', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_inv_warehouse_usetype'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_inv_warehouse_usetype t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

INSERT INTO tbl_setup_valuation_method (code, name)
OUTPUT N'tbl_setup_valuation_method', inserted.code, inserted.name INTO #added
SELECT i.code, i.name
FROM #incoming i
WHERE i.table_name = N'tbl_setup_valuation_method'
  AND NOT EXISTS (
        SELECT 1 FROM tbl_setup_valuation_method t
        WHERE LTRIM(RTRIM(LOWER(t.code))) = LTRIM(RTRIM(LOWER(i.code)))
           OR LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name))));

-- What happened ----------------------------------------------------------------
SELECT i.table_name,
       COUNT(*)                                        AS in_file,
       SUM(CASE WHEN a.code IS NULL THEN 0 ELSE 1 END) AS inserted,
       SUM(CASE WHEN a.code IS NULL THEN 1 ELSE 0 END) AS already_there
FROM #incoming i
LEFT JOIN #added a
       ON a.table_name = i.table_name AND a.code = i.code AND a.name = i.name
GROUP BY i.table_name
ORDER BY i.table_name;

SELECT 'inserted' AS state, table_name, code, name FROM #added ORDER BY table_name, name;

-- Where the file and the table disagree: the name is already there under a different
-- code. Nothing is changed for these - they are listed so somebody can decide.
SELECT N'tbl_setup_item_brand' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_brand t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_brand'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_class' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_class t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_class'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_material' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_material t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_material'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_name' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_name t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_name'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_pump_count' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_pump_count t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_pump_count'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_pump_type' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_pump_type t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_pump_type'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_type' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_type t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_type'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_item_unit_measurement' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_item_unit_measurement t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_item_unit_measurement'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_payment_terms' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_payment_terms t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_payment_terms'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_social' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_social t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_social'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_bpi_entity' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_bpi_entity t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_bpi_entity'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_bpi_industries' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_bpi_industries t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_bpi_industries'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_position' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_position t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_position'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_inv_warehouse_usetype' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_inv_warehouse_usetype t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_inv_warehouse_usetype'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
UNION ALL
SELECT N'tbl_setup_valuation_method' AS table_name, i.code AS file_code, t.code AS existing_code, t.name
FROM #incoming i JOIN tbl_setup_valuation_method t
       ON LTRIM(RTRIM(LOWER(t.name))) = LTRIM(RTRIM(LOWER(i.name)))
WHERE i.table_name = N'tbl_setup_valuation_method'
  AND LTRIM(RTRIM(LOWER(t.code))) <> LTRIM(RTRIM(LOWER(i.code)))
ORDER BY table_name, name;

SELECT 'also appears in Access Control as a position' AS note, code, name
FROM #added WHERE table_name = N'tbl_position' ORDER BY name;

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

IF OBJECT_ID('tempdb..#incoming') IS NOT NULL DROP TABLE #incoming;
IF OBJECT_ID('tempdb..#added') IS NOT NULL DROP TABLE #added;
