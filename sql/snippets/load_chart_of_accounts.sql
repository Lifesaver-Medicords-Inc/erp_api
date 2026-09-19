-- Replaces the whole chart of accounts (tbl_setup_chart_of_accounts) with the 226 accounts
-- of 'GL ACCOUNTS - CLASSIFIED.xlsx', tab '142 training'. One transaction:
--   1. adds a Chart Class for any of ASSET / LIABILITY / EQUITY / REVENUE / EXPENSE missing
--   2. DELETES every existing account (logged as DELETE in z_tbl_setup_chart_of_accounts_at)
--   3. inserts the accounts; a sub-account gets its parent as GROUP and GROUP_ID
--   4. keeps the row ids the posting code looks up directly - NEVER renumber these:
--        40030  2000001-1  ACCOUNTS PAYABLE - IR, bulk IR, PV, supplier CM, DM
--        50030  5000021-2  NON-TRADE EXPENSE - supplier CM debit, DM credit
--        70032  1000002-1  TRADE RECEIVABLE - SI, PAYR, customer CM
--        70034  1000001-5  CASH ON HAND - cash PAYR / PV, cash-flow report
--        70035  CASH       CASH ON BANK - non-cash PAYR / PV, cash-flow report
--        70037  4000001-5  SALES - SI, customer CM
--        70038  2000008-2  ADVANCE PAYMENT - PAYR / PV differences
--   5. checks the result and lists rows elsewhere still pointing at a deleted account
-- Nothing is kept unless every check passes.
--
-- HOW TO RUN: open in SSMS on the target database (or sqlcmd -d <db> -i <this file> -b).
-- With @commit = 0 everything is done and counted, then rolled back. Set @commit = 1 to keep
-- it. After a commit, restart that database's API (or wait up to an hour): the chart is cached
-- in Redis and Accounting keeps showing the old one until the cache is cleared.
--
-- 14 rows break Chart Of Accounts Setup's code rules; they load as they are, but that
-- screen will refuse to re-save them until corrected:
--   row 39: CASH CASH - ASSET codes should start with 10000
--   row 40: 1000001-1 CHINA BANK - does not start with its group's code CASH
--   row 41: 1000001-2 METROBANK 1 - does not start with its group's code CASH
--   row 42: 1000001-2.1 METRO BANK 2 - does not start with its group's code CASH
--   row 43: 1000001-3 BANCO DE ORO 1 - does not start with its group's code CASH
--   row 44: 1000001-3.1 BANCO DE ORO 2 - does not start with its group's code CASH
--   row 45: 1000001-4 PETTY CASH - does not start with its group's code CASH
--   row 46: 1000001-5 CASH ON HAND - does not start with its group's code CASH
--   row 47: 1000001-6 EAST WEST BANK - does not start with its group's code CASH
--   row 48: 1000001-6.1 EAST WEST BANK 2 - does not start with its group's code CASH
--   row 49: 1000001-7 SECURITY BANK - does not start with its group's code CASH
--   row 72: 14000014-2 PETTY CASH GAIN - REVENUE codes should start with 40000
--   row 73: 14000015-1 FOREIGN EXCHANGE GAIN/LOSS - REVENUE codes should start with 40000
--   row 84: 15000014-3 PETTY CASH LOSS - EXPENSE codes should start with 50000

SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;  -- 1 = keep the changes, 0 = trial run

BEGIN TRANSACTION;

DECLARE @at_date nvarchar(19) = CONVERT(nvarchar(19), GETDATE(), 120);
DECLARE @at_machine nvarchar(100) = HOST_NAME();

-- 1. One class per type. An existing class is reused; a missing type gets one.
DECLARE @types TABLE (type nvarchar(20) PRIMARY KEY, code nvarchar(20) NOT NULL);
INSERT INTO @types (type, code) VALUES
    (N'ASSET', N'10000'),
    (N'LIABILITY', N'20000'),
    (N'EQUITY', N'30000'),
    (N'REVENUE', N'40000'),
    (N'EXPENSE', N'50000');

INSERT INTO tbl_setup_chart_class (type, code, name)
SELECT t.type, t.code, t.type FROM @types t
WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_chart_class c WHERE UPPER(LTRIM(RTRIM(c.type))) = t.type);
PRINT CONCAT('Classes created: ', @@ROWCOUNT);

DECLARE @class TABLE (type nvarchar(20) PRIMARY KEY, id bigint NOT NULL, name nvarchar(max) NULL);
INSERT INTO @class (type, id, name)
SELECT t.type, c.id, c.name
FROM @types t
CROSS APPLY (SELECT TOP 1 id, name FROM tbl_setup_chart_class x
             WHERE UPPER(LTRIM(RTRIM(x.type))) = t.type ORDER BY x.id) c;
IF (SELECT COUNT(*) FROM @class) <> 5
BEGIN
    RAISERROR('A class type is still missing from tbl_setup_chart_class - nothing was changed.', 16, 1);
    ROLLBACK TRANSACTION; RETURN;
END
SELECT type AS class_type, id AS class_id, name AS class_name FROM @class ORDER BY type;

-- 2. The accounts on the tab, in tab order.
DECLARE @coa TABLE (seq int PRIMARY KEY, code nvarchar(100) NOT NULL UNIQUE, name nvarchar(300) NOT NULL,
                    type nvarchar(20) NOT NULL, parent_code nvarchar(100) NULL, keep_id bigint NULL, new_id bigint NULL);
INSERT INTO @coa (seq, code, name, type, parent_code, keep_id) VALUES
    (1, N'1000002-1', N'ACCOUNTS RECEIVABLES-TRADE PRIVATE', N'ASSET', NULL, 70032),
    (2, N'1000003-1', N'ACCOUNTS RECEIVABLES-TRADE GOVERNMENT', N'ASSET', NULL, NULL),
    (3, N'1000004', N'ADVANCES TO EMPLOYEE', N'ASSET', NULL, NULL),
    (4, N'1000004-1', N'ADV SALVACION LIZA RAMOS', N'ASSET', N'1000004', NULL),
    (5, N'1000004-2', N'ADV RONALIZA BETCO', N'ASSET', N'1000004', NULL),
    (6, N'1000004-3', N'ADV PAMELA SIMPLICIO', N'ASSET', N'1000004', NULL),
    (7, N'1000004-4', N'ADV AYDE SORIANO', N'ASSET', N'1000004', NULL),
    (8, N'1000004-5', N'ADV RONA CASILAN', N'ASSET', N'1000004', NULL),
    (9, N'1000004-6', N'ADV EFREN ESPALDON', N'ASSET', N'1000004', NULL),
    (10, N'1000004-7', N'ADV MARIO MORA', N'ASSET', N'1000004', NULL),
    (11, N'1000004-8', N'ADV ROLANDO BALDIVINO', N'ASSET', N'1000004', NULL),
    (12, N'1000004-9', N'ADV ROLANTE NAQUILA', N'ASSET', N'1000004', NULL),
    (13, N'1000004-10', N'ADV JULIE CRISTY CESTONA', N'ASSET', N'1000004', NULL),
    (14, N'1000004-11', N'ADV RIZALINA GAJISAN', N'ASSET', N'1000004', NULL),
    (15, N'1000004-12', N'ADV VIDELIO CALAGUI', N'ASSET', N'1000004', NULL),
    (16, N'1000004-13', N'ADV ROMEL LUBIANO', N'ASSET', N'1000004', NULL),
    (17, N'1000004-14', N'ADV MARVEN LUZON', N'ASSET', N'1000004', NULL),
    (18, N'1000005-1', N'MERCHANDISE INVENTORY', N'ASSET', NULL, NULL),
    (19, N'1000006-1', N'VAT INPUT', N'ASSET', NULL, NULL),
    (20, N'1000007-1', N'CREDITABLE WITHHOLDING TAX', N'ASSET', NULL, NULL),
    (21, N'1000007-2', N'CREDITABLE WITHHOLDING TAX (PENDING)', N'ASSET', NULL, NULL),
    (22, N'1000008-1', N'REALTY UNITS', N'ASSET', NULL, NULL),
    (23, N'1000009-1', N'FURNITURE & FIXTURE', N'ASSET', NULL, NULL),
    (24, N'1000010-1', N'OFFICE EQUIPTMENT', N'ASSET', NULL, NULL),
    (25, N'1000011-1', N'VEHICLE & DELIVERY EQUIPMENT', N'ASSET', NULL, NULL),
    (26, N'1000012-1', N'ACCUM.DEP-REALTY UNITS', N'ASSET', NULL, NULL),
    (27, N'1000012-2', N'ACCUM.DEP-FUR & FIXTURES', N'ASSET', NULL, NULL),
    (28, N'1000012-3', N'ACCUM.DEP-OFFICE EQUIPMENT', N'ASSET', NULL, NULL),
    (29, N'1000012-4', N'ACCUM.DEP-VEHICLES', N'ASSET', NULL, NULL),
    (30, N'1000013-1', N'GENERAL INVENTORY', N'ASSET', NULL, NULL),
    (31, N'1000014-1', N'PETTY CASH FUND', N'ASSET', NULL, NULL),
    (32, N'1000014-4', N'PETTY CASH FUND (PENDING)', N'ASSET', NULL, NULL),
    (33, N'1000015-1', N'AR UNAPPLIED', N'ASSET', NULL, NULL),
    (34, N'1000016-1', N'FOR TESTING ITEMS', N'ASSET', NULL, NULL),
    (35, N'1000018', N'SOFTWARE SYSTEM', N'ASSET', NULL, NULL),
    (36, N'1000018-01', N'Q-ERP', N'ASSET', N'1000018', NULL),
    (37, N'1000018-02', N'WEBSITE', N'ASSET', N'1000018', NULL),
    (38, N'CASH', N'CASH', N'ASSET', NULL, 70035),
    (39, N'1000001-1', N'CHINA BANK', N'ASSET', N'CASH', NULL),
    (40, N'1000001-2', N'METROBANK 1', N'ASSET', N'CASH', NULL),
    (41, N'1000001-2.1', N'METRO BANK 2', N'ASSET', N'CASH', NULL),
    (42, N'1000001-3', N'BANCO DE ORO 1', N'ASSET', N'CASH', NULL),
    (43, N'1000001-3.1', N'BANCO DE ORO 2', N'ASSET', N'CASH', NULL),
    (44, N'1000001-4', N'PETTY CASH', N'ASSET', N'CASH', NULL),
    (45, N'1000001-5', N'CASH ON HAND', N'ASSET', N'CASH', 70034),
    (46, N'1000001-6', N'EAST WEST BANK', N'ASSET', N'CASH', NULL),
    (47, N'1000001-6.1', N'EAST WEST BANK 2', N'ASSET', N'CASH', NULL),
    (48, N'1000001-7', N'SECURITY BANK', N'ASSET', N'CASH', NULL),
    (49, N'2000001-1', N'ACCOUNTS PAYABLE-TRADE', N'LIABILITY', NULL, 40030),
    (50, N'2000001-2', N'ACCRUED EXPENSE PAYABLE', N'LIABILITY', NULL, NULL),
    (51, N'2000002-1', N'PHILHEALTH PREMIUM PAYABLE', N'LIABILITY', NULL, NULL),
    (52, N'2000003-1', N'HDMF PREMIUM PAYABLE', N'LIABILITY', NULL, NULL),
    (53, N'2000003-2', N'PAG-IBIG LOAN PAYABLE', N'LIABILITY', NULL, NULL),
    (54, N'2000003-3', N'HDMF LOAN PAYABLE', N'LIABILITY', NULL, NULL),
    (55, N'2000004-1', N'SSS PREMIUM PAYABLE', N'LIABILITY', NULL, NULL),
    (56, N'2000004-2', N'SSS LOAN PAYABLE', N'LIABILITY', NULL, NULL),
    (57, N'2000005-1', N'VAT OUTPUT', N'LIABILITY', NULL, NULL),
    (58, N'2000005-2', N'VAT PAYABLE', N'LIABILITY', NULL, NULL),
    (59, N'2000006-1', N'EXPANDED WITHHOLDING TAX', N'LIABILITY', NULL, NULL),
    (60, N'2000006-2', N'WITHHOLDING TAX PAYABLE-COMP', N'LIABILITY', NULL, NULL),
    (61, N'2000007-1', N'LOAN PAYABLE', N'LIABILITY', NULL, NULL),
    (62, N'2000008-1', N'ACCOUNTS PAYABLE-NON-TRADE', N'LIABILITY', NULL, NULL),
    (63, N'2000008-2', N'ADVANCE PAYABLE', N'LIABILITY', NULL, 70038),
    (64, N'2000009-1', N'INCOME TAX PAYABLE', N'LIABILITY', NULL, NULL),
    (65, N'3000001-1', N'RETAINED EARNINGS', N'EQUITY', NULL, NULL),
    (66, N'3000001-2', N'RESTRICTED RETAINED EARNINGS', N'EQUITY', NULL, NULL),
    (67, N'3000002-1', N'SHARE CAPITAL', N'EQUITY', NULL, NULL),
    (68, N'14000014-2', N'PETTY CASH GAIN', N'REVENUE', NULL, NULL),
    (69, N'14000015-1', N'FOREIGN EXCHANGE GAIN/LOSS', N'REVENUE', NULL, NULL),
    (70, N'4000001-1', N'SALES-VATABLE', N'REVENUE', NULL, NULL),
    (71, N'4000001-2', N'SALES-GOVERNMENT', N'REVENUE', NULL, NULL),
    (72, N'4000001-3', N'SALES-ZERO RATED', N'REVENUE', NULL, NULL),
    (73, N'4000001-4', N'SALES-DISCOUNT', N'REVENUE', NULL, NULL),
    (74, N'4000001-5', N'SALES', N'REVENUE', NULL, 70037),
    (75, N'4000006-1', N'CHARGES DISCOUNT', N'REVENUE', NULL, NULL),
    (76, N'4000006-2', N'PURCHASE DISCOUNT', N'REVENUE', NULL, NULL),
    (77, N'4000007-1', N'MISCELLANEOUS INCOME', N'REVENUE', NULL, NULL),
    (78, N'4000008-1', N'INVENTORY GAIN OR LOSS', N'REVENUE', NULL, NULL),
    (79, N'15000014-3', N'PETTY CASH LOSS', N'EXPENSE', NULL, NULL),
    (80, N'5000001-1', N'ADVERTISING & PROMOTION', N'EXPENSE', NULL, NULL),
    (81, N'5000002-1', N'COMMISSION', N'EXPENSE', NULL, NULL),
    (82, N'5000003', N'COMMUNICATION, LIGHT & WATER', N'EXPENSE', NULL, NULL),
    (83, N'5000003-1', N'MERALCO 343612970101', N'EXPENSE', N'5000003', NULL),
    (84, N'5000003-2', N'GLOBE 917-6746088', N'EXPENSE', N'5000003', NULL),
    (85, N'5000003-3', N'PLDT 832-9723', N'EXPENSE', N'5000003', NULL),
    (86, N'5000003-4', N'GLOBE 917-8966066', N'EXPENSE', N'5000003', NULL),
    (87, N'5000003-5', N'GLOBE 917-8966071', N'EXPENSE', N'5000003', NULL),
    (88, N'5000003-6', N'GLOBE 917-8966070', N'EXPENSE', N'5000003', NULL),
    (89, N'5000003-7', N'SUN 922-8966067', N'EXPENSE', N'5000003', NULL),
    (90, N'5000003-8', N'SUN 922-8966071', N'EXPENSE', N'5000003', NULL),
    (91, N'5000003-9', N'SUN 922-8966074', N'EXPENSE', N'5000003', NULL),
    (92, N'5000003-10', N'SUN 922-8966065', N'EXPENSE', N'5000003', NULL),
    (93, N'5000003-11', N'SUN 922-8966070', N'EXPENSE', N'5000003', NULL),
    (94, N'5000003-12', N'SUN 922-8966066', N'EXPENSE', N'5000003', NULL),
    (95, N'5000003-13', N'MAYNILAD BILL', N'EXPENSE', N'5000003', NULL),
    (96, N'5000003-14', N'PLDT 247-7015', N'EXPENSE', N'5000003', NULL),
    (97, N'5000003-15', N'PLDT 247-7016', N'EXPENSE', N'5000003', NULL),
    (98, N'5000003-16', N'PLDT 247-7017', N'EXPENSE', N'5000003', NULL),
    (99, N'5000003-17', N'PLDT 245-0287', N'EXPENSE', N'5000003', NULL),
    (100, N'5000003-18', N'PLDT 245-0277', N'EXPENSE', N'5000003', NULL),
    (101, N'5000003-19', N'PLDT 245-0279', N'EXPENSE', N'5000003', NULL),
    (102, N'5000003-20', N'PLDT 247-6895', N'EXPENSE', N'5000003', NULL),
    (103, N'5000003-21', N'MERALCO 343612930101', N'EXPENSE', N'5000003', NULL),
    (104, N'5000003-22', N'MERALCO 343613190101', N'EXPENSE', N'5000003', NULL),
    (105, N'5000003-23', N'CONVERGE ICT', N'EXPENSE', N'5000003', NULL),
    (106, N'5000004', N'DEPRECIATION', N'EXPENSE', NULL, NULL),
    (107, N'5000004-01', N'REALTY UNIT', N'EXPENSE', N'5000004', NULL),
    (108, N'5000004-02', N'FURNITURE & FIXTURES', N'EXPENSE', N'5000004', NULL),
    (109, N'5000004-03', N'OFFICE EQUIPMENT', N'EXPENSE', N'5000004', NULL),
    (110, N'5000004-04', N'VEHICLES', N'EXPENSE', N'5000004', NULL),
    (111, N'5000005', N'FRINGE BENEFITS (COLA, HOLIDAY PAY, AL)', N'EXPENSE', NULL, NULL),
    (112, N'5000005-01', N'COLA', N'EXPENSE', N'5000005', NULL),
    (113, N'5000005-02', N'HOLIDAY PAY', N'EXPENSE', N'5000005', NULL),
    (114, N'5000005-03', N'ALLOWANCE', N'EXPENSE', N'5000005', NULL),
    (115, N'5000006', N'FUEL & OIL', N'EXPENSE', NULL, NULL),
    (116, N'5000006-1', N'DIESEL PZI-839', N'EXPENSE', N'5000006', NULL),
    (117, N'5000006-2', N'DIESEL TSH-475', N'EXPENSE', N'5000006', NULL),
    (118, N'5000006-3', N'DIESEL PMK-364', N'EXPENSE', N'5000006', NULL),
    (119, N'5000006-4', N'DIESEL UUF-288', N'EXPENSE', N'5000006', NULL),
    (120, N'5000006-5', N'DIESEL RLY-683', N'EXPENSE', N'5000006', NULL),
    (121, N'5000006-6', N'GASOLINE QY-1062', N'EXPENSE', N'5000006', NULL),
    (122, N'5000006-7', N'GASOLINE NM-5667', N'EXPENSE', N'5000006', NULL),
    (123, N'5000006-8', N'OTHERS', N'EXPENSE', N'5000006', NULL),
    (124, N'5000006-9', N'DIESEL CO-6589', N'EXPENSE', N'5000006', NULL),
    (125, N'5000006-10', N'DIESEL WKI-771', N'EXPENSE', N'5000006', NULL),
    (126, N'5000006-11', N'DIESEL ZTT-705', N'EXPENSE', N'5000006', NULL),
    (127, N'5000006-12', N'DIESEL LONG BROTHER', N'EXPENSE', N'5000006', NULL),
    (128, N'5000006-13', N'DIESEL FOR UJU-518', N'EXPENSE', N'5000006', NULL),
    (129, N'5000007', N'OFFICE SUPPLIES', N'EXPENSE', NULL, NULL),
    (130, N'5000007-01', N'OFFICE SUPPLIES', N'EXPENSE', N'5000007', NULL),
    (131, N'5000008', N'REPAIRS & MAINTENANCE-LABOR & MATERIALS', N'EXPENSE', NULL, NULL),
    (132, N'5000008-2', N'CO-6589 ISUZU MUX', N'EXPENSE', N'5000008', NULL),
    (133, N'5000008-3', N'WKI-771 MITSUBISHI MONTERO', N'EXPENSE', N'5000008', NULL),
    (134, N'5000008-4', N'ZTT-705 TOYOTA FORTUNER', N'EXPENSE', N'5000008', NULL),
    (135, N'5000008-5', N'UJU-518 TALL BROTHER', N'EXPENSE', N'5000008', NULL),
    (136, N'5000008-6', N'RLY-683 BIG BROTHER', N'EXPENSE', N'5000008', NULL),
    (137, N'5000008-7', N'TSH-475 OLD BROTHER', N'EXPENSE', N'5000008', NULL),
    (138, N'5000008-8', N'RMK-364 CUTE', N'EXPENSE', N'5000008', NULL),
    (139, N'5000008-9', N'UUF-288 L-300', N'EXPENSE', N'5000008', NULL),
    (140, N'5000008-10', N'PZI-839 HINO', N'EXPENSE', N'5000008', NULL),
    (141, N'5000008-11', N'QY-1062', N'EXPENSE', N'5000008', NULL),
    (142, N'5000008-12', N'NM-5667', N'EXPENSE', N'5000008', NULL),
    (143, N'5000008-13', N'OFFICE EQUIPMENT', N'EXPENSE', N'5000008', NULL),
    (144, N'5000008-14', N'LONG BROTHER', N'EXPENSE', N'5000008', NULL),
    (145, N'5000008-1', N'INTEREST AND PENALTIES', N'EXPENSE', NULL, NULL),
    (146, N'5000009', N'REPRESENTATION EXPENSES', N'EXPENSE', NULL, NULL),
    (147, N'5000009-01', N'MEAL ALLOWANCE', N'EXPENSE', N'5000009', NULL),
    (148, N'5000009-1', N'PURCHASES', N'EXPENSE', NULL, NULL),
    (149, N'5000010', N'SALARIES AND ALLOWANCE', N'EXPENSE', NULL, NULL),
    (150, N'5000010-2', N'SAL RIZALINA GAJISAN', N'EXPENSE', N'5000010', NULL),
    (151, N'5000010-3', N'SAL VIDELIO CALAGUI', N'EXPENSE', N'5000010', NULL),
    (152, N'5000010-4', N'SAL RONALIZA BETCO', N'EXPENSE', N'5000010', NULL),
    (153, N'5000010-5', N'SAL PAMELA SIMPLICIO', N'EXPENSE', N'5000010', NULL),
    (154, N'5000010-6', N'SAL AYDE SORIANO', N'EXPENSE', N'5000010', NULL),
    (155, N'5000010-7', N'SAL RONA CASILAN', N'EXPENSE', N'5000010', NULL),
    (156, N'5000010-8', N'SAL EFREN ESPALDON', N'EXPENSE', N'5000010', NULL),
    (157, N'5000010-9', N'SAL MARIO MORA', N'EXPENSE', N'5000010', NULL),
    (158, N'5000010-10', N'SAL SALVACION LIZA RAMOS', N'EXPENSE', N'5000010', NULL),
    (159, N'5000010-11', N'SAL JULIE CRISTY CESTONA', N'EXPENSE', N'5000010', NULL),
    (160, N'5000010-12', N'SAL ROLANDO BALDIVINO', N'EXPENSE', N'5000010', NULL),
    (161, N'5000010-13', N'SAL ROLANTE NAQUILA', N'EXPENSE', N'5000010', NULL),
    (162, N'5000010-14', N'SAL ROMEL LUBIANO', N'EXPENSE', N'5000010', NULL),
    (163, N'5000010-15', N'SAL BRYAN GABRIEL TAN', N'EXPENSE', N'5000010', NULL),
    (164, N'5000010-16', N'SAL JOVY BRILLANTES', N'EXPENSE', N'5000010', NULL),
    (165, N'5000010-17', N'SAL ARTEMIO CABILA', N'EXPENSE', N'5000010', NULL),
    (166, N'5000010-18', N'SAL EDUARDO FERRANCO', N'EXPENSE', N'5000010', NULL),
    (167, N'5000010-19', N'SALARY FOR AGENCY', N'EXPENSE', N'5000010', NULL),
    (168, N'5000010-20', N'SAL CECILYN ORIAN', N'EXPENSE', N'5000010', NULL),
    (169, N'5000010-21', N'SALARIES & WAGES', N'EXPENSE', N'5000010', NULL),
    (170, N'5000010-1', N'MISCELLANEOUS', N'EXPENSE', NULL, NULL),
    (171, N'5000011', N'OVERTIME PAY', N'EXPENSE', NULL, NULL),
    (172, N'5000012', N'13TH MONTH PAY', N'EXPENSE', NULL, NULL),
    (173, N'5000012-1', N'OFFICERS BENEFIT', N'EXPENSE', NULL, NULL),
    (174, N'5000013', N'SSS, HDMF, PHILHEALTH', N'EXPENSE', NULL, NULL),
    (175, N'5000013-1', N'SSS PREMIUM EXPENSE', N'EXPENSE', N'5000013', NULL),
    (176, N'5000013-2', N'PHILHEALTH PREMIUM EXPENSE', N'EXPENSE', N'5000013', NULL),
    (177, N'5000013-3', N'HDMF EXPENSE', N'EXPENSE', N'5000013', NULL),
    (178, N'5000014', N'TAXES & LICENSES', N'EXPENSE', NULL, NULL),
    (179, N'5000014-01', N'TAXES & LICENSES', N'EXPENSE', N'5000014', NULL),
    (180, N'5000015', N'FINES/ PENALTIES', N'EXPENSE', NULL, NULL),
    (181, N'5000015-1', N'CHECKPOINT', N'EXPENSE', N'5000015', NULL),
    (182, N'5000015-2', N'FINES & PENALTIES', N'EXPENSE', N'5000015', NULL),
    (183, N'5000016', N'TRANSPORTATION, TOLL, & PARKING', N'EXPENSE', NULL, NULL),
    (184, N'5000016-01', N'TRANSPORTATION', N'EXPENSE', N'5000016', NULL),
    (185, N'5000016-02', N'TOLL FEE', N'EXPENSE', N'5000016', NULL),
    (186, N'5000016-03', N'PARKING FOR UUF-288', N'EXPENSE', N'5000016', NULL),
    (187, N'5000016-04', N'PARKING FOR RMK-364', N'EXPENSE', N'5000016', NULL),
    (188, N'5000016-05', N'PARKING FOR PZI-839', N'EXPENSE', N'5000016', NULL),
    (189, N'5000016-06', N'PARKING FOR QY-1062', N'EXPENSE', N'5000016', NULL),
    (190, N'5000016-07', N'PARKING FOR NM-5667', N'EXPENSE', N'5000016', NULL),
    (191, N'5000016-08', N'PARKING FOR WKI-771', N'EXPENSE', N'5000016', NULL),
    (192, N'5000016-09', N'PARKING FOR ISUZU MUX', N'EXPENSE', N'5000016', NULL),
    (193, N'5000017', N'VEHICLE REGISTRATION FEES', N'EXPENSE', NULL, NULL),
    (194, N'5000017-01', N'LTO REG WKI-771', N'EXPENSE', N'5000017', NULL),
    (195, N'5000017-02', N'LTO REG QY-1062', N'EXPENSE', N'5000017', NULL),
    (196, N'5000017-03', N'LTO REG RLY-683', N'EXPENSE', N'5000017', NULL),
    (197, N'5000017-04', N'LTO REG RMK-364', N'EXPENSE', N'5000017', NULL),
    (198, N'5000017-05', N'LTO REG TSH-475', N'EXPENSE', N'5000017', NULL),
    (199, N'5000017-06', N'LTO REG NM-5667', N'EXPENSE', N'5000017', NULL),
    (200, N'5000017-07', N'LTO REG UJU-518', N'EXPENSE', N'5000017', NULL),
    (201, N'5000017-08', N'LTO REG UUF-288', N'EXPENSE', N'5000017', NULL),
    (202, N'5000017-09', N'LTO REG PZI-839', N'EXPENSE', N'5000017', NULL),
    (203, N'5000018', N'SHIPPING, FREIGHT HANDLING, BROKERS FEE', N'EXPENSE', NULL, NULL),
    (204, N'5000018-1', N'SHIPPING', N'EXPENSE', N'5000018', NULL),
    (205, N'5000018-2', N'FREIGHT AND HANDLING', N'EXPENSE', N'5000018', NULL),
    (206, N'5000018-3', N'WAREHOUSING', N'EXPENSE', N'5000018', NULL),
    (207, N'5000018-4', N'BROKER''S FEE', N'EXPENSE', N'5000018', NULL),
    (208, N'5000019', N'INSURANCE', N'EXPENSE', NULL, NULL),
    (209, N'5000019-2', N'WKI-771', N'EXPENSE', N'5000019', NULL),
    (210, N'5000019-3', N'CO-6589', N'EXPENSE', N'5000019', NULL),
    (211, N'5000019-4', N'ZTT-705', N'EXPENSE', N'5000019', NULL),
    (212, N'5000019-5', N'RMK-364', N'EXPENSE', N'5000019', NULL),
    (213, N'5000019-6', N'RLY-683', N'EXPENSE', N'5000019', NULL),
    (214, N'5000019-7', N'TSH-475', N'EXPENSE', N'5000019', NULL),
    (215, N'5000019-8', N'PZI-839', N'EXPENSE', N'5000019', NULL),
    (216, N'5000019-9', N'UJU-518', N'EXPENSE', N'5000019', NULL),
    (217, N'5000019-10', N'UUF-288 L-300', N'EXPENSE', N'5000019', NULL),
    (218, N'5000019-11', N'6750 CONTRACTORS ALL RISK', N'EXPENSE', N'5000019', NULL),
    (219, N'5000019-12', N'LONG BROTHER', N'EXPENSE', N'5000019', NULL),
    (220, N'5000019-13', N'2017 VOLVO V60CC D4', N'EXPENSE', N'5000019', NULL),
    (221, N'5000019-14', N'NM-5667', N'EXPENSE', N'5000019', NULL),
    (222, N'5000019-1', N'BANK CHARGES', N'EXPENSE', NULL, NULL),
    (223, N'5000020-1', N'INVENTORY LOSS', N'EXPENSE', NULL, NULL),
    (224, N'5000021-1', N'TAX OTHERS', N'EXPENSE', NULL, NULL),
    (225, N'5000021-2', N'NON-TRADE INVENTORY', N'EXPENSE', NULL, 50030),
    (226, N'5000022-1', N'JANITORIAL AND MESSENGERIAL EXPENSE', N'EXPENSE', NULL, NULL);

-- 3. The old chart, recorded before it goes.
SELECT COUNT(*) AS accounts_before FROM tbl_setup_chart_of_accounts;
INSERT INTO z_tbl_setup_chart_of_accounts_at (ref_id, code, name, account_class, class_id, [group], group_id,
       cash_flow_category, liquidity_class, AT_ACTION, AT_DATE, AT_USER, AT_USER_ID, IP_ADDRESS, MACHINE_NAME, MOTHERBOARD_SERIAL_NO)
SELECT id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class,
       N'DELETE', @at_date, N'chart of accounts load', N'', N'', @at_machine, N''
FROM tbl_setup_chart_of_accounts;
DELETE FROM tbl_setup_chart_of_accounts;
PRINT CONCAT('Accounts deleted: ', @@ROWCOUNT);

-- 4. The new chart. The accounts the posting code reaches by id keep that id.
SET IDENTITY_INSERT tbl_setup_chart_of_accounts ON;
INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class)
SELECT a.keep_id, a.code, a.name, c.name, c.id, N'', 0, N'', N''
FROM @coa a JOIN @class c ON c.type = a.type
WHERE a.keep_id IS NOT NULL;
SET IDENTITY_INSERT tbl_setup_chart_of_accounts OFF;

INSERT INTO tbl_setup_chart_of_accounts (code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class)
SELECT a.code, a.name, c.name, c.id, N'', 0, N'', N''
FROM @coa a JOIN @class c ON c.type = a.type
WHERE a.keep_id IS NULL
ORDER BY a.seq;

UPDATE a SET new_id = t.id FROM @coa a JOIN tbl_setup_chart_of_accounts t ON t.code = a.code;

-- A sub-account's GROUP is its parent's name (what Chart Of Accounts Setup saves) and
-- GROUP_ID its parent's id (what that screen reads to show the group).
UPDATE t SET [group] = p.name, group_id = p.new_id
FROM tbl_setup_chart_of_accounts t
JOIN @coa a ON a.new_id = t.id
JOIN @coa p ON p.code = a.parent_code;

INSERT INTO z_tbl_setup_chart_of_accounts_at (ref_id, code, name, account_class, class_id, [group], group_id,
       cash_flow_category, liquidity_class, AT_ACTION, AT_DATE, AT_USER, AT_USER_ID, IP_ADDRESS, MACHINE_NAME, MOTHERBOARD_SERIAL_NO)
SELECT id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class,
       N'INSERT', @at_date, N'chart of accounts load', N'', N'', @at_machine, N''
FROM tbl_setup_chart_of_accounts;

-- 5. Checks. Any failure undoes everything.
IF (SELECT COUNT(*) FROM tbl_setup_chart_of_accounts) <> 226
BEGIN
    RAISERROR('The chart does not hold exactly 226 accounts - nothing was changed.', 16, 1);
    ROLLBACK TRANSACTION; RETURN;
END
IF EXISTS (SELECT 1 FROM @coa WHERE new_id IS NULL)
BEGIN
    RAISERROR('An account was not inserted - nothing was changed.', 16, 1);
    ROLLBACK TRANSACTION; RETURN;
END
IF EXISTS (SELECT 1 FROM @coa WHERE keep_id IS NOT NULL AND new_id <> keep_id)
BEGIN
    RAISERROR('An account did not keep the id the posting code needs - nothing was changed.', 16, 1);
    ROLLBACK TRANSACTION; RETURN;
END
IF EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts t JOIN @coa a ON a.new_id = t.id
           WHERE a.parent_code IS NOT NULL AND ISNULL(t.group_id, 0) = 0)
BEGIN
    RAISERROR('A sub-account has no group - nothing was changed.', 16, 1);
    ROLLBACK TRANSACTION; RETURN;
END

SELECT t.id AS kept_id, t.code, t.name, t.account_class, v.used_for
FROM tbl_setup_chart_of_accounts t
JOIN (VALUES
    (40030, N'ACCOUNTS PAYABLE - IR, bulk IR, PV, supplier CM, DM'),
    (50030, N'NON-TRADE EXPENSE - supplier CM debit, DM credit'),
    (70032, N'TRADE RECEIVABLE - SI, PAYR, customer CM'),
    (70034, N'CASH ON HAND - cash PAYR / PV, cash-flow report'),
    (70035, N'CASH ON BANK - non-cash PAYR / PV, cash-flow report'),
    (70037, N'SALES - SI, customer CM'),
    (70038, N'ADVANCE PAYMENT - PAYR / PV differences')
) v (id, used_for) ON v.id = t.id
ORDER BY t.id;

SELECT c.type AS class_type, COUNT(*) AS accounts,
       SUM(CASE WHEN t.group_id <> 0 THEN 1 ELSE 0 END) AS sub_accounts
FROM tbl_setup_chart_of_accounts t JOIN @class c ON c.id = t.class_id
GROUP BY c.type ORDER BY c.type;

-- 6. Still pointing at an account that is gone. Reported only - none of these are changed.
SELECT 'tbl_setup_tax' AS points_from, COUNT(*) AS rows_pointing_at_no_account FROM tbl_setup_tax
  WHERE (ISNULL(coa_purchase_id, 0) <> 0 AND coa_purchase_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts))
     OR (ISNULL(coa_sales_id, 0) <> 0 AND coa_sales_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts))
UNION ALL SELECT 'tbl_bpi_finance.finance_account_id', COUNT(*) FROM tbl_bpi_finance
  WHERE ISNULL(finance_account_id, 0) <> 0 AND finance_account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)
UNION ALL SELECT 'tbl_bpi_items.item_account_id', COUNT(*) FROM tbl_bpi_items
  WHERE ISNULL(item_account_id, 0) <> 0 AND item_account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)
UNION ALL SELECT 'tbl_accounting_bulk_invoice_receipt_details.account_id', COUNT(*) FROM tbl_accounting_bulk_invoice_receipt_details
  WHERE ISNULL(account_id, 0) <> 0 AND account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)
UNION ALL SELECT 'tbl_accounting_journal_entry_details.posting_ref_id', COUNT(*) FROM tbl_accounting_journal_entry_details
  WHERE ISNULL(posting_ref_id, 0) <> 0 AND posting_ref_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts);

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'COMMITTED';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'TRIAL RUN - rolled back. Set @commit = 1 to keep the changes.';
END
