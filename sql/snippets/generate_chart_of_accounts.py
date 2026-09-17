"""Builds the chart-of-accounts replacement (SQL) from the '142 training' tab of
'GL ACCOUNTS - CLASSIFIED.xlsx'.

The tab is a two-level list:

  column A        a section: ASSET, LIABILITIES, EQUITY, REVENUE, EXPENSE
  columns B + D   an account (code, name) - a few rows carry the name in C instead of D
  columns C + E   a sub-account (code, name) of the account above it

The SQL this writes, in one transaction:

  1. makes sure tbl_setup_chart_class has a class for each of the five types - an existing
     class is reused (lowest id per type), a missing type gets one (code 30000 / 50000 ...)
  2. deletes EVERY row of tbl_setup_chart_of_accounts
  3. inserts the tab's accounts; a sub-account gets its parent as GROUP (name) and GROUP_ID
  4. gives seven accounts the row ids the posting code looks up directly (KEEP_IDS below),
     so Invoice Receipt, Bulk IR, Sales Invoice, Payment Receipt, Payment Voucher and the
     memos keep posting
  5. writes a DELETE and an INSERT row per account to z_tbl_setup_chart_of_accounts_at
  6. checks the result and lists what still points at an account that no longer exists
     (tax setup, BPI, bulk IR lines, journal lines) - it reports these, it changes none

Normally run through load_chart_of_accounts.ps1. On its own:

  python generate_chart_of_accounts.py "<GL ACCOUNTS - CLASSIFIED.xlsx>" [output.sql] [--commit]

Without --commit the SQL is a trial run that rolls itself back.
"""

import os
import sys

import openpyxl

SHEET = '142 training'

# Section label on the sheet -> class type used by the reports and Chart Class Setup, plus the
# code and name given to a class this load has to create.
SECTIONS = {
    'ASSET': ('ASSET', '10000'),
    'LIABILITIES': ('LIABILITY', '20000'),
    'EQUITY': ('EQUITY', '30000'),
    'REVENUE': ('REVENUE', '40000'),
    'EXPENSE': ('EXPENSE', '50000'),
}

# Row ids compiled into the posting code, and the new account that takes each one
# (decided by the user 2026-09-17). Until the GL mapping layer (spec 12.8) replaces
# these ids, an account that loses its id stops those documents from posting.
KEEP_IDS = {
    40030: ('2000001-1', 'ACCOUNTS PAYABLE - IR, bulk IR, PV, supplier CM, DM'),
    50030: ('5000021-2', 'NON-TRADE EXPENSE - supplier CM debit, DM credit'),
    70032: ('1000002-1', 'TRADE RECEIVABLE - SI, PAYR, customer CM'),
    70034: ('1000001-5', 'CASH ON HAND - cash PAYR / PV, cash-flow report'),
    70035: ('CASH', 'CASH ON BANK - non-cash PAYR / PV, cash-flow report'),
    70037: ('4000001-5', 'SALES - SI, customer CM'),
    70038: ('2000008-2', 'ADVANCE PAYMENT - PAYR / PV differences'),
}

EXPECTED_COUNT = 226


def text(value):
    """Cell value as the code or name it shows: 1000004.0 -> '1000004'."""
    if value is None:
        return ''
    if isinstance(value, float) and value.is_integer():
        value = int(value)
    return str(value).strip()


def sql_str(value):
    return "N'" + value.replace("'", "''") + "'"


def read_accounts(path):
    wb = openpyxl.load_workbook(path, read_only=True, data_only=True)
    if SHEET not in wb.sheetnames:
        raise SystemExit(f"No '{SHEET}' tab in {path} (tabs: {', '.join(wb.sheetnames)})")

    accounts, errors = [], []
    section, parent = None, None
    for number, row in enumerate(wb[SHEET].iter_rows(values_only=True), 1):
        a, b, c, d, e = (list(row) + [None] * 5)[:5]
        if text(a):
            label = text(a).upper()
            if label not in SECTIONS:
                errors.append(f'row {number}: unknown section {text(a)!r}')
            section, parent = label, None
            continue
        if text(b):
            # Rows 134 and 136 of the tab carry the account name in C, not D.
            name = text(d) or text(c)
            account = {'row': number, 'section': section, 'code': text(b), 'name': name, 'parent': None}
            accounts.append(account)
            parent = account
            continue
        if text(c):
            if parent is None:
                errors.append(f'row {number}: sub-account {text(c)} has no account above it')
            accounts.append({'row': number, 'section': section, 'code': text(c), 'name': text(e),
                             'parent': parent['code'] if parent else None})
            continue
        if any(v is not None and text(v) for v in row):
            errors.append(f'row {number}: not an account or a section: {row}')

    return accounts, errors


def check(accounts):
    errors, warnings = [], []

    if len(accounts) != EXPECTED_COUNT:
        errors.append(f'{len(accounts)} accounts on the tab, expected {EXPECTED_COUNT}')

    seen = {}
    for account in accounts:
        if account['section'] is None:
            errors.append(f"row {account['row']}: account {account['code']} is above every section")
        if not account['name']:
            errors.append(f"row {account['row']}: account {account['code']} has no name")
        if account['code'] in seen:
            errors.append(f"row {account['row']}: code {account['code']} repeats row {seen[account['code']]}")
        seen[account['code']] = account['row']

    for keep_id, (code, use) in KEEP_IDS.items():
        if code not in seen:
            errors.append(f'id {keep_id} ({use}) is meant for {code}, which is not on the tab')

    # Chart Of Accounts Setup's own rules. The tab is management's list, so rows that break
    # them are loaded as they are and listed - but that screen will refuse to re-save them.
    by_code = {a['code']: a for a in accounts}
    for account in accounts:
        if account['section'] not in SECTIONS:
            continue
        if account['parent'] is None:
            prefix = SECTIONS.get(account['section'], ('', ''))[1]
            if not account['code'].startswith(prefix) or len(account['code']) < 6:
                warnings.append(f"row {account['row']}: {account['code']} {account['name']} - "
                                f"{SECTIONS[account['section']][0]} codes should start with {prefix}")
        else:
            parent = by_code[account['parent']]
            if not account['code'].startswith(parent['code']):
                warnings.append(f"row {account['row']}: {account['code']} {account['name']} - does not "
                                f"start with its group's code {parent['code']}")

    return errors, warnings


def build_sql(accounts, commit):
    keep_by_code = {code: keep_id for keep_id, (code, _) in KEEP_IDS.items()}
    out = []
    w = out.append

    w('SET NOCOUNT ON;')
    w('SET XACT_ABORT ON;')
    w('')
    w('BEGIN TRANSACTION;')
    w('')
    w("DECLARE @at_date nvarchar(19) = CONVERT(nvarchar(19), GETDATE(), 120);")
    w("DECLARE @at_machine nvarchar(100) = HOST_NAME();")
    w('')
    w('-- 1. One class per type. An existing class is reused; a missing type gets one.')
    w('DECLARE @types TABLE (type nvarchar(20) PRIMARY KEY, code nvarchar(20) NOT NULL);')
    w('INSERT INTO @types (type, code) VALUES')
    w(',\n'.join(f'    ({sql_str(t)}, {sql_str(p)})' for t, p in SECTIONS.values()) + ';')
    w('')
    w('INSERT INTO tbl_setup_chart_class (type, code, name)')
    w('SELECT t.type, t.code, t.type FROM @types t')
    w('WHERE NOT EXISTS (SELECT 1 FROM tbl_setup_chart_class c WHERE UPPER(LTRIM(RTRIM(c.type))) = t.type);')
    w("PRINT CONCAT('Classes created: ', @@ROWCOUNT);")
    w('')
    w('DECLARE @class TABLE (type nvarchar(20) PRIMARY KEY, id bigint NOT NULL, name nvarchar(max) NULL);')
    w('INSERT INTO @class (type, id, name)')
    w('SELECT t.type, c.id, c.name')
    w('FROM @types t')
    w('CROSS APPLY (SELECT TOP 1 id, name FROM tbl_setup_chart_class x')
    w('             WHERE UPPER(LTRIM(RTRIM(x.type))) = t.type ORDER BY x.id) c;')
    w('IF (SELECT COUNT(*) FROM @class) <> 5')
    w("    RAISERROR('A class type is still missing from tbl_setup_chart_class - nothing was changed.', 16, 1);")
    w("SELECT type AS class_type, id AS class_id, name AS class_name FROM @class ORDER BY type;")
    w('')
    w('-- 2. The accounts on the tab, in tab order.')
    w('DECLARE @coa TABLE (seq int PRIMARY KEY, code nvarchar(100) NOT NULL UNIQUE, name nvarchar(300) NOT NULL,')
    w('                    type nvarchar(20) NOT NULL, parent_code nvarchar(100) NULL, keep_id bigint NULL, new_id bigint NULL);')
    w('INSERT INTO @coa (seq, code, name, type, parent_code, keep_id) VALUES')
    rows = []
    for seq, account in enumerate(accounts, 1):
        parent = sql_str(account['parent']) if account['parent'] else 'NULL'
        keep = str(keep_by_code[account['code']]) if account['code'] in keep_by_code else 'NULL'
        rows.append(f"    ({seq}, {sql_str(account['code'])}, {sql_str(account['name'])}, "
                    f"{sql_str(SECTIONS[account['section']][0])}, {parent}, {keep})")
    w(',\n'.join(rows) + ';')
    w('')
    w('-- 3. The old chart, recorded before it goes.')
    w("SELECT COUNT(*) AS accounts_before FROM tbl_setup_chart_of_accounts;")
    w('INSERT INTO z_tbl_setup_chart_of_accounts_at (ref_id, code, name, account_class, class_id, [group], group_id,')
    w('       cash_flow_category, liquidity_class, AT_ACTION, AT_DATE, AT_USER, AT_USER_ID, IP_ADDRESS, MACHINE_NAME, MOTHERBOARD_SERIAL_NO)')
    w("SELECT id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class,")
    w("       N'DELETE', @at_date, N'chart of accounts load', N'', N'', @at_machine, N''")
    w('FROM tbl_setup_chart_of_accounts;')
    w('DELETE FROM tbl_setup_chart_of_accounts;')
    w("PRINT CONCAT('Accounts deleted: ', @@ROWCOUNT);")
    w('')
    w('-- 4. The new chart. The accounts the posting code reaches by id keep that id.')
    w('SET IDENTITY_INSERT tbl_setup_chart_of_accounts ON;')
    w('INSERT INTO tbl_setup_chart_of_accounts (id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class)')
    w("SELECT a.keep_id, a.code, a.name, c.name, c.id, N'', 0, N'', N''")
    w('FROM @coa a JOIN @class c ON c.type = a.type')
    w('WHERE a.keep_id IS NOT NULL;')
    w('SET IDENTITY_INSERT tbl_setup_chart_of_accounts OFF;')
    w('')
    w('INSERT INTO tbl_setup_chart_of_accounts (code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class)')
    w("SELECT a.code, a.name, c.name, c.id, N'', 0, N'', N''")
    w('FROM @coa a JOIN @class c ON c.type = a.type')
    w('WHERE a.keep_id IS NULL')
    w('ORDER BY a.seq;')
    w('')
    w('UPDATE a SET new_id = t.id FROM @coa a JOIN tbl_setup_chart_of_accounts t ON t.code = a.code;')
    w('')
    w("-- A sub-account's GROUP is its parent's name (what Chart Of Accounts Setup saves) and")
    w('-- GROUP_ID its parent\'s id (what that screen reads to show the group).')
    w('UPDATE t SET [group] = p.name, group_id = p.new_id')
    w('FROM tbl_setup_chart_of_accounts t')
    w('JOIN @coa a ON a.new_id = t.id')
    w('JOIN @coa p ON p.code = a.parent_code;')
    w('')
    w('INSERT INTO z_tbl_setup_chart_of_accounts_at (ref_id, code, name, account_class, class_id, [group], group_id,')
    w('       cash_flow_category, liquidity_class, AT_ACTION, AT_DATE, AT_USER, AT_USER_ID, IP_ADDRESS, MACHINE_NAME, MOTHERBOARD_SERIAL_NO)')
    w("SELECT id, code, name, account_class, class_id, [group], group_id, cash_flow_category, liquidity_class,")
    w("       N'INSERT', @at_date, N'chart of accounts load', N'', N'', @at_machine, N''")
    w('FROM tbl_setup_chart_of_accounts;')
    w('')
    w('-- 5. Checks. Any failure undoes everything.')
    w(f'IF (SELECT COUNT(*) FROM tbl_setup_chart_of_accounts) <> {len(accounts)}')
    w(f"    RAISERROR('The chart does not hold exactly {len(accounts)} accounts - nothing was changed.', 16, 1);")
    w('IF EXISTS (SELECT 1 FROM @coa WHERE new_id IS NULL)')
    w("    RAISERROR('An account was not inserted - nothing was changed.', 16, 1);")
    w('IF EXISTS (SELECT 1 FROM @coa WHERE keep_id IS NOT NULL AND new_id <> keep_id)')
    w("    RAISERROR('An account did not keep the id the posting code needs - nothing was changed.', 16, 1);")
    w('IF EXISTS (SELECT 1 FROM tbl_setup_chart_of_accounts t JOIN @coa a ON a.new_id = t.id')
    w('           WHERE a.parent_code IS NOT NULL AND ISNULL(t.group_id, 0) = 0)')
    w("    RAISERROR('A sub-account has no group - nothing was changed.', 16, 1);")
    w('')
    w('SELECT t.id AS kept_id, t.code, t.name, t.account_class, v.used_for')
    w('FROM tbl_setup_chart_of_accounts t')
    w('JOIN (VALUES')
    w(',\n'.join(f'    ({keep_id}, {sql_str(use)})' for keep_id, (_, use) in KEEP_IDS.items()))
    w(') v (id, used_for) ON v.id = t.id')
    w('ORDER BY t.id;')
    w('')
    w('SELECT c.type AS class_type, COUNT(*) AS accounts,')
    w('       SUM(CASE WHEN t.group_id <> 0 THEN 1 ELSE 0 END) AS sub_accounts')
    w('FROM tbl_setup_chart_of_accounts t JOIN @class c ON c.id = t.class_id')
    w('GROUP BY c.type ORDER BY c.type;')
    w('')
    w('-- 6. Still pointing at an account that is gone. Reported only - none of these are changed.')
    w("SELECT 'tbl_setup_tax' AS points_from, COUNT(*) AS rows_pointing_at_no_account FROM tbl_setup_tax")
    w('  WHERE (ISNULL(coa_purchase_id, 0) <> 0 AND coa_purchase_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts))')
    w('     OR (ISNULL(coa_sales_id, 0) <> 0 AND coa_sales_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts))')
    w("UNION ALL SELECT 'tbl_bpi_finance.finance_account_id', COUNT(*) FROM tbl_bpi_finance")
    w('  WHERE ISNULL(finance_account_id, 0) <> 0 AND finance_account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)')
    w("UNION ALL SELECT 'tbl_bpi_items.item_account_id', COUNT(*) FROM tbl_bpi_items")
    w('  WHERE ISNULL(item_account_id, 0) <> 0 AND item_account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)')
    w("UNION ALL SELECT 'tbl_accounting_bulk_invoice_receipt_details.account_id', COUNT(*) FROM tbl_accounting_bulk_invoice_receipt_details")
    w('  WHERE ISNULL(account_id, 0) <> 0 AND account_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts)')
    w("UNION ALL SELECT 'tbl_accounting_journal_entry_details.posting_ref_id', COUNT(*) FROM tbl_accounting_journal_entry_details")
    w('  WHERE ISNULL(posting_ref_id, 0) <> 0 AND posting_ref_id NOT IN (SELECT id FROM tbl_setup_chart_of_accounts);')
    w('')
    if commit:
        w('COMMIT TRANSACTION;')
        w("PRINT 'COMMITTED';")
    else:
        w('ROLLBACK TRANSACTION;')
        w("PRINT 'ROLLED BACK - trial run, nothing was kept';")
    w('')
    return '\n'.join(out)


def main(argv):
    args = [a for a in argv if not a.startswith('--')]
    commit = '--commit' in argv
    if not args:
        raise SystemExit(__doc__)
    workbook = args[0]
    output = args[1] if len(args) > 1 else None

    if not os.path.isfile(workbook):
        raise SystemExit(f'Workbook not found: {workbook}')
    lock = os.path.join(os.path.dirname(os.path.abspath(workbook)), '~$' + os.path.basename(workbook))
    if os.path.exists(lock):
        print(f'NOTE: the workbook is open in Excel ({os.path.basename(lock)} exists). '
              'Only what was last SAVED is loaded.')

    accounts, errors = read_accounts(workbook)
    more_errors, warnings = check(accounts)
    errors += more_errors

    tops = sum(1 for a in accounts if a['parent'] is None)
    print(f"'{SHEET}': {len(accounts)} accounts ({tops} accounts, {len(accounts) - tops} sub-accounts)")
    for section, (class_type, _) in SECTIONS.items():
        print(f'  {class_type:<10} {sum(1 for a in accounts if a["section"] == section)}')
    if warnings:
        print(f'{len(warnings)} row(s) break Chart Of Accounts Setup\'s code rules and are loaded as they are:')
        for warning in warnings:
            print('  ' + warning)
    if errors:
        print('The tab failed its checks - nothing was written:')
        for error in errors:
            print('  ' + error)
        return 1

    sql = build_sql(accounts, commit)
    if output:
        with open(output, 'w', encoding='utf-8-sig') as handle:
            handle.write(sql)
    else:
        sys.stdout.write(sql)
    return 0


if __name__ == '__main__':
    sys.exit(main(sys.argv[1:]))
