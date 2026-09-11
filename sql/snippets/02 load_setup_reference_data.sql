-- ============================================================
-- Lightspeed ERP - setup reference data
-- Generated from 'data migration excel format/ERP SETUP_2.xlsx'
-- Target: SQL Server 2012.
--
-- HOW TO RUN
--     sqlcmd -S <server> -U sa -P <pw> -d <database> -b -I -i <this file>
--     The -I matters: QUOTED_IDENTIFIER must be ON.
--
-- Run AFTER the API has started once against the database, so AutoMigrate
-- has created the tables. This loads data only - it creates nothing.
--
-- IDEMPOTENT, and additive only. Each row is inserted only when its CODE is
-- not already present, so re-running changes nothing and a value someone
-- edited by hand is never overwritten. Nothing is ever deleted: this file
-- tops a database up to the workbook, it does not make it match.
--
-- The workbook's 'Position' sheet is deliberately NOT loaded. It holds BPI
-- CONTACT job titles (Contractor, Owner, Project Manager...), while
-- tbl_position holds EMPLOYEE positions that drive access control (Admin,
-- Warehouse, Sales Representatives...). They are different lists that happen
-- to share a name, and tbl_bpi_contacts.position is free text with no lookup
-- table behind it - so there is nowhere to put that sheet without inventing
-- a destination. See the report that accompanied this file.
-- ============================================================
SET NOCOUNT ON;
GO

-- ---------- Item Brand -> tbl_setup_item_brand (12 rows) ----------
-- Acejet was originally loaded as ACJ; the agreed code is ACE. Rename the
-- EXISTING row rather than inserting ACE alongside it, which would leave two
-- Acejet brands. Guarded both ways so re-running is safe and so it cannot
-- collide with an ACE row that already exists.
IF OBJECT_ID('dbo.tbl_setup_item_brand') IS NOT NULL
   AND EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'ACJ')
   AND NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'ACE')
    UPDATE [dbo].[tbl_setup_item_brand] SET code = N'ACE' WHERE code = N'ACJ';
GO

IF OBJECT_ID('dbo.tbl_setup_item_brand') IS NULL
    PRINT 'SKIPPED tbl_setup_item_brand - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'ACE')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'ACE', N'Acejet');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'CLP')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'CLP', N'Calpeda');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'DAN')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'DAN', N'Danfoss');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'DEL')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'DEL', N'Delta');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'DYNA')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'DYNA', N'Dynaflo');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'ETN')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'ETN', N'EATON');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'FTI')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'FTI', N'Finish Thompson Inc.');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'ISP')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'ISP', N'Instapure');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'LGI')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'LGI', N'Little Giant');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'NB')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'NB', N'No Brand');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'PTA')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'PTA', N'Pentair Aurora');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_brand] WHERE code = N'SHN')
        INSERT INTO [dbo].[tbl_setup_item_brand] (code, name) VALUES (N'SHN', N'Shinmaywa');
END
GO

-- ---------- Item Class -> tbl_setup_item_class (12 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_class') IS NULL
    PRINT 'SKIPPED tbl_setup_item_class - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'TAI')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'TAI', N'TOOLS AND INSTRUMENTS');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'CON')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'CON', N'CONSTRUCTION');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'ELEC')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'ELEC', N'ELECTRICAL');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'EXP')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'EXP', N'EXPENSE (EMPTY SPECS)');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'HW')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'HW', N'HARDWARE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'IMP')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'IMP', N'IMPORT PURCHASE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'LOC')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'LOC', N'LOCAL PURCHASE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'OFF')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'OFF', N'OFFICE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'PMB')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'PMB', N'PLUMBING');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'SER')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'SER', N'SERVICES');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'WM')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'WM', N'WATER METER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_class] WHERE code = N'PMPP')
        INSERT INTO [dbo].[tbl_setup_item_class] (code, name) VALUES (N'PMPP', N'PUMP PARTS');
END
GO

-- ---------- Item Material -> tbl_setup_item_material (10 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_material') IS NULL
    PRINT 'SKIPPED tbl_setup_item_material - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'AISL 304')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'AISL 304', N'STAINLESS STEEL 304');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'AISL 316')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'AISL 316', N'STAINLESS STEEL 316');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'DI')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'DI', N'DUCTILE IRON');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'BI')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'BI', N'BLACK IRON');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'GB')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'GB', N'BRONZE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'GG')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'GG', N'CAST IRON');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'GI')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'GI', N'GALVANIZED IRON');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'GO')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'GO', N'BRASS');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'PPO-GF20')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'PPO-GF20', N'NORYL');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_material] WHERE code = N'FS')
        INSERT INTO [dbo].[tbl_setup_item_material] (code, name) VALUES (N'FS', N'FORGED STEEL');
END
GO

-- ---------- Item Name -> tbl_setup_item_name (35 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_name') IS NULL
    PRINT 'SKIPPED tbl_setup_item_name - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'ARV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'ARV', N'AIR RELEASE VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'BV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'BV', N'BALL VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'BT')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'BT', N'BLADDER TANK');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'BFV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'BFV', N'BUTTERFLY VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'CP')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'CP', N'COMMON PACKAGE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'CONS')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'CONS', N'CONSUMABLES');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'CTRL')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'CTRL', N'CONTROLLER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'DCH')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'DCH', N'DISCHARGE COMMON HEADER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'DRB')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'DRB', N'DISCHARGE RUBBER BELLOW');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'ELEC')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'ELEC', N'ELECTRICAL');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'FS')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'FS', N'FLOAT SWITCH');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'FLV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'FLV', N'FLOAT VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'FM')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'FM', N'FLOW METER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'FTV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'FTV', N'FOOT VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'GV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'GV', N'GATE VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'HDW')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'HDW', N'HARDWARE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'LLC')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'LLC', N'LIQUID LEVEL CONTROL');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'OTH')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'OTH', N'OTHER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PNT')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PNT', N'PAINT');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PG')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PG', N'PRESSURE GAUGE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PRV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PRV', N'PRESSURE RELIEF VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PS')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PS', N'PRESSURE SWITCH');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PT')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PT', N'PRESSURE TRANSDUCER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PMP')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PMP', N'PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'PMPP')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'PMPP', N'PUMP PART');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'SER')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'SER', N'SERVICE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'SCV')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'SCV', N'SUCTION CHECK VALVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'SCH')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'SCH', N'SUCTION COMMON HEADER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'SRB')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'SRB', N'SUCTION RUBBER BELLOW');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'SUP')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'SUP', N'SUPPPLIES');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'TAI')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'TAI', N'TOOLS AND INSTRUMENTS');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'VFD')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'VFD', N'VARIABLE FREQUENCY DRIVE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'WC')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'WC', N'WASTE CONE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'WM')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'WM', N'WATER METER');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_name] WHERE code = N'WMAT')
        INSERT INTO [dbo].[tbl_setup_item_name] (code, name) VALUES (N'WMAT', N'WIRING MATERIALS');
END
GO

-- ---------- Item Pump Count -> tbl_setup_item_pump_count (5 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_pump_count') IS NULL
    PRINT 'SKIPPED tbl_setup_item_pump_count - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_count] WHERE code = N'SMP')
        INSERT INTO [dbo].[tbl_setup_item_pump_count] (code, name) VALUES (N'SMP', N'Simplex');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_count] WHERE code = N'DPX')
        INSERT INTO [dbo].[tbl_setup_item_pump_count] (code, name) VALUES (N'DPX', N'Duplex');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_count] WHERE code = N'TPX')
        INSERT INTO [dbo].[tbl_setup_item_pump_count] (code, name) VALUES (N'TPX', N'Triplex');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_count] WHERE code = N'QPX')
        INSERT INTO [dbo].[tbl_setup_item_pump_count] (code, name) VALUES (N'QPX', N'Quadruplex');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_count] WHERE code = N'QPX2')
        INSERT INTO [dbo].[tbl_setup_item_pump_count] (code, name) VALUES (N'QPX2', N'Quintuplex');
END
GO

-- ---------- Item Pump Type -> tbl_setup_item_pump_type (7 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_pump_type') IS NULL
    PRINT 'SKIPPED tbl_setup_item_pump_type - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'BP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'BP', N'BOOSTER PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'CPS')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'CPS', N'CPS');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'DP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'DP', N'DRAINAGE PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'NUFP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'NUFP', N'NON-ULFM FIRE PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'NUJP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'NUJP', N'NON-ULFM JOCKEY PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'SP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'SP', N'SEWAGE PUMP');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_pump_type] WHERE code = N'TP')
        INSERT INTO [dbo].[tbl_setup_item_pump_type] (code, name) VALUES (N'TP', N'TRANSFER PUMP');
END
GO

-- ---------- Item Type -> tbl_setup_item_type (2 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_type') IS NULL
    PRINT 'SKIPPED tbl_setup_item_type - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_type] WHERE code = N'TRD')
        INSERT INTO [dbo].[tbl_setup_item_type] (code, name) VALUES (N'TRD', N'TRADE');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_type] WHERE code = N'NTR')
        INSERT INTO [dbo].[tbl_setup_item_type] (code, name) VALUES (N'NTR', N'NON-TRADE');
END
GO

-- ---------- Unit of Measure -> tbl_setup_item_unit_measurement (29 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_item_unit_measurement') IS NULL
    PRINT 'SKIPPED tbl_setup_item_unit_measurement - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'BAG')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'BAG', N'Bag');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'BKS')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'BKS', N'Books');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'BTL')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'BTL', N'Bottle');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'CAN')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'CAN', N'Can');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'CBY')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'CBY', N'Cubic Yard');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'DOZ.')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'DOZ.', N'Dozen');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'DRM')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'DRM', N'Drum');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'FEET')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'FEET', N'Feet');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'GAL')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'GAL', N'Gallon');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'GR.')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'GR.', N'Gross');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'KEG')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'KEG', N'Keg');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'KG')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'KG', N'Kilogram');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'LM')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'LM', N'Linear Meter');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'LOT')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'LOT', N'Lot');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'LTR')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'LTR', N'Liter');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'MTR')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'MTR', N'Meter');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'PACK')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'PACK', N'Pack');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'PAD')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'PAD', N'Pad');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'PAIL')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'PAIL', N'Pail');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'PC')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'PC', N'Piece');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'PRS')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'PRS', N'Pairs');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'QAT')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'QAT', N'Quart');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'REAM')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'REAM', N'Ream');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'ROLL')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'ROLL', N'Roll');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'SET')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'SET', N'Set');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'SHEET')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'SHEET', N'Sheet');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'TANK')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'TANK', N'Tank');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'TIN')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'TIN', N'Tin');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_item_unit_measurement] WHERE code = N'UNIT')
        INSERT INTO [dbo].[tbl_setup_item_unit_measurement] (code, name) VALUES (N'UNIT', N'Unit');
END
GO

-- ---------- Payment terms -> tbl_setup_payment_terms (13 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_payment_terms') IS NULL
    PRINT 'SKIPPED tbl_setup_payment_terms - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'DP50')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'DP50', N'Downpayment 50%');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'DP25')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'DP25', N'Downpayment 25%');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'DP75')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'DP75', N'Downpayment 75%');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'COD')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'COD', N'Cash on Delivery');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'COD-C')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'COD-C', N'Cash on Delivery - Cheque');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'NET15')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'NET15', N'Net 15');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'NET30')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'NET30', N'Net 30');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'NET60')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'NET60', N'Net 60');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'PDC30')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'PDC30', N'Post Dated Cheque 30');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'PDC45')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'PDC45', N'Post Dated Cheque 45');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'PDC60')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'PDC60', N'Post Dated Cheque 60');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'50-50')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'50-50', N'50%-50%');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_payment_terms] WHERE code = N'30-20-50')
        INSERT INTO [dbo].[tbl_setup_payment_terms] (code, name) VALUES (N'30-20-50', N'30%- 20%- 50%');
END
GO

-- ---------- Socials -> tbl_setup_social (12 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_social') IS NULL
    PRINT 'SKIPPED tbl_setup_social - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'EML')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'EML', N'Email');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'FB')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'FB', N'Facebook');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'IG')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'IG', N'Instagram');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'INT')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'INT', N'Internet Search');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'LNK')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'LNK', N'LinkedIn');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'REF')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'REF', N'Referral');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'ST')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'ST', N'Sales Team');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'TG')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'TG', N'Telegram');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'TK')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'TK', N'TikTok');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'VBR')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'VBR', N'Viber');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'WEB')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'WEB', N'Website');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_social] WHERE code = N'WSP')
        INSERT INTO [dbo].[tbl_setup_social] (code, name) VALUES (N'WSP', N'WhatsApp');
END
GO

-- ---------- Entity Type -> tbl_setup_bpi_entity (7 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_bpi_entity') IS NULL
    PRINT 'SKIPPED tbl_setup_bpi_entity - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'AFF')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'AFF', N'Affiliated');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'BLK')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'BLK', N'Blacklisted');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'CLD')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'CLD', N'Closed');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'CUS')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'CUS', N'Customer');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'NAF')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'NAF', N'Non-Affiliated');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'SUP')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'SUP', N'Supplier');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_entity] WHERE code = N'TSP')
        INSERT INTO [dbo].[tbl_setup_bpi_entity] (code, name) VALUES (N'TSP', N'Temporary Supplier');
END
GO

-- ---------- Industries -> tbl_setup_bpi_industries (26 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_bpi_industries') IS NULL
    PRINT 'SKIPPED tbl_setup_bpi_industries - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'AGRI')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'AGRI', N'Agriculture/ Irrigation');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'AQUA')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'AQUA', N'Aquaculture');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'CHEM')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'CHEM', N'Chemical');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'COLD')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'COLD', N'Cold Storage');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'CON')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'CON', N'Construction');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'DATA')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'DATA', N'Data Centers');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'DOM')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'DOM', N'Domestic/ Residential');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'FILTER')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'FILTER', N'Filtration');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'FIRE')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'FIRE', N'Fire Systems');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'FBV')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'FBV', N'Food / Beverage');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'GRND')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'GRND', N'Groundwater');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'HC')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'HC', N'Healthcare');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'HVAC')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'HVAC', N'HVAC');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'IND')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'IND', N'Industrial');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'ICE')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'ICE', N'Ice Plant');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'LAND')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'LAND', N'Landscaping');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'MRN')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'MRN', N'Marine');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'MINE')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'MINE', N'Mining');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'MUNI')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'MUNI', N'Municipal');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'OIL')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'OIL', N'Oil / Gas');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'PHRM')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'PHRM', N'Pharmaceuticals');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'PWR')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'PWR', N'Power Generation');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'POOL')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'POOL', N'Pools');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'SC')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'SC', N'Semiconductors');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'SKCM')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'SKCM', N'Skyscrapers/ Commercial');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'WT')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'WT', N'Water Treatment');
	IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_bpi_industries] WHERE code = N'OTH')
        INSERT INTO [dbo].[tbl_setup_bpi_industries] (code, name) VALUES (N'OTH', N'Others');
END
GO

-- ---------- Warehouse Use Type -> tbl_inv_warehouse_usetype (5 rows) ----------
IF OBJECT_ID('dbo.tbl_inv_warehouse_usetype') IS NULL
    PRINT 'SKIPPED tbl_inv_warehouse_usetype - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_inv_warehouse_usetype] WHERE code = N'DMG')
        INSERT INTO [dbo].[tbl_inv_warehouse_usetype] (code, name) VALUES (N'DMG', N'DAMAGED ITEMS');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_inv_warehouse_usetype] WHERE code = N'GEN')
        INSERT INTO [dbo].[tbl_inv_warehouse_usetype] (code, name) VALUES (N'GEN', N'GENERAL');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_inv_warehouse_usetype] WHERE code = N'OUT')
        INSERT INTO [dbo].[tbl_inv_warehouse_usetype] (code, name) VALUES (N'OUT', N'OUTBOUND');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_inv_warehouse_usetype] WHERE code = N'REC')
        INSERT INTO [dbo].[tbl_inv_warehouse_usetype] (code, name) VALUES (N'REC', N'RECEIVING');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_inv_warehouse_usetype] WHERE code = N'SP')
        INSERT INTO [dbo].[tbl_inv_warehouse_usetype] (code, name) VALUES (N'SP', N'SPARE PARTS');
END
GO

-- ---------- Valuation Method -> tbl_setup_valuation_method (6 rows) ----------
IF OBJECT_ID('dbo.tbl_setup_valuation_method') IS NULL
    PRINT 'SKIPPED tbl_setup_valuation_method - table does not exist';
ELSE
BEGIN
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'FIFO')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'FIFO', N'First In First Out');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'LIFO')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'LIFO', N'Last In First Out');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'WAC')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'WAC', N'Weighted Average Cost');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'MAC')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'MAC', N'Moving Average Cost');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'JIT')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'JIT', N'Just In Time');
    IF NOT EXISTS (SELECT 1 FROM [dbo].[tbl_setup_valuation_method] WHERE code = N'STD')
        INSERT INTO [dbo].[tbl_setup_valuation_method] (code, name) VALUES (N'STD', N'Standard Costing');
END
GO

-- ---------- summary ----------
IF OBJECT_ID('dbo.tbl_setup_item_brand') IS NOT NULL
    SELECT 'tbl_setup_item_brand' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_brand];
IF OBJECT_ID('dbo.tbl_setup_item_class') IS NOT NULL
    SELECT 'tbl_setup_item_class' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_class];
IF OBJECT_ID('dbo.tbl_setup_item_material') IS NOT NULL
    SELECT 'tbl_setup_item_material' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_material];
IF OBJECT_ID('dbo.tbl_setup_item_name') IS NOT NULL
    SELECT 'tbl_setup_item_name' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_name];
IF OBJECT_ID('dbo.tbl_setup_item_pump_count') IS NOT NULL
    SELECT 'tbl_setup_item_pump_count' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_pump_count];
IF OBJECT_ID('dbo.tbl_setup_item_pump_type') IS NOT NULL
    SELECT 'tbl_setup_item_pump_type' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_pump_type];
IF OBJECT_ID('dbo.tbl_setup_item_type') IS NOT NULL
    SELECT 'tbl_setup_item_type' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_type];
IF OBJECT_ID('dbo.tbl_setup_item_unit_measurement') IS NOT NULL
    SELECT 'tbl_setup_item_unit_measurement' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_item_unit_measurement];
IF OBJECT_ID('dbo.tbl_setup_payment_terms') IS NOT NULL
    SELECT 'tbl_setup_payment_terms' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_payment_terms];
IF OBJECT_ID('dbo.tbl_setup_social') IS NOT NULL
    SELECT 'tbl_setup_social' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_social];
IF OBJECT_ID('dbo.tbl_setup_bpi_entity') IS NOT NULL
    SELECT 'tbl_setup_bpi_entity' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_bpi_entity];
IF OBJECT_ID('dbo.tbl_setup_bpi_industries') IS NOT NULL
    SELECT 'tbl_setup_bpi_industries' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_bpi_industries];
IF OBJECT_ID('dbo.tbl_inv_warehouse_usetype') IS NOT NULL
    SELECT 'tbl_inv_warehouse_usetype' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_inv_warehouse_usetype];
IF OBJECT_ID('dbo.tbl_setup_valuation_method') IS NOT NULL
    SELECT 'tbl_setup_valuation_method' AS [table], COUNT(*) AS rows_now FROM [dbo].[tbl_setup_valuation_method];
GO
