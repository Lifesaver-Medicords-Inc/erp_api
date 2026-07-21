-- Restores these chart-of-accounts rows with their original explicit ids.
-- id is an identity column, so IDENTITY_INSERT must be ON to write explicit
-- values into it. Each row is guarded with IF NOT EXISTS so this is safe to
-- re-run without throwing a duplicate-key error on rows that are already there.

SET IDENTITY_INSERT tbl_setup_chart_of_accounts ON;

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 40029)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (40029, '500008', 'INVENTORY LOSS', 'EXPENSES', 3079, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 40030)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (40030, '200001', 'ACCOUNTS PAYABLE', 'LIABILITIES', 3080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 50030)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (50030, '500009', 'NON-TRADE EXPENSE', 'EXPENSES', 3079, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 60030)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (60030, '100003', 'AR UNAPPLIED', 'ASSETS', 4080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70032)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70032, '100002', 'TRADE RECEIVABLE', 'ASSETS', 4080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70033)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70033, '400006', 'INVENTORY GAIN', 'REVENUES', 5080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70034)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70034, '100001-01', 'CASH ON HAND', 'ASSETS', 4080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70035)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70035, '100001-02', 'CASH ON BANK', 'ASSETS', 4080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70036)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70036, '200002', 'ACCRUED EXPENSE PAYABLE', 'LIABILITIES', 3080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70037)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70037, '400001', 'SALES', 'REVENUES', 5080, '', 0);

IF NOT EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts WHERE id = 70038)
    INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id)
    VALUES (70038, '200003', 'ADVANCE PAYMENT', 'LIABILITIES', 3080, '', 0);

SET IDENTITY_INSERT tbl_setup_chart_of_accounts OFF;
