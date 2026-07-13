ALTER VIEW [dbo].[get_customer] AS
SELECT *
FROM (
        SELECT a.id AS based_id,
            a.id AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(a.name, '') + ',' + ISNULL(a.main_tel_no, '') + ',' + ISNULL(a.main_website, '') + ',' + ISNULL(a.tin, '') AS edit_history
        FROM z_tbl_bpi_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
        UNION ALL
        SELECT a.id,
            '' AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(a.branch_name, '') + ', ' + ISNULL(a.transaction_type, '') + ', ' + ISNULL(a.class_name, '') + ', ' + ISNULL(a.branch_tel_no, '') + ', ' + ISNULL(a.branch_website, '') + ', ' + ISNULL(a.customer_code, '') + ', ' + ISNULL(c.code, '') + ', ' + ISNULL(a.supplier_code, '') + ', ' + ISNULL(a.fax_no, '') + ', ' + ISNULL(a.notes, '') AS edit_history
        FROM z_tbl_bpi_general_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_social c ON a.social = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(a.name, '') + ', ' + ISNULL(a.number, '') + ', ' + ISNULL(a.name, '') + ', ' + ISNULL(a.email, '') + ', ' + ISNULL(a.preferences, '') + ', ' + ISNULL(c.code, '') AS edit_history
        FROM z_tbl_bpi_contacts_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_social c ON a.position = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(a.location, '') + ' ' AS edit_history
        FROM z_tbl_bpi_address_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
        UNION ALL
        SELECT a.finance_based_id AS based_id,
            a.finance_branch_id AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(CAST(a.finance_account_id AS NVARCHAR), '') + ', ' + ISNULL(c.code, '') AS edit_history
        FROM z_tbl_bpi_finance_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_payment_terms c ON a.finance_payment_terms_id = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(a.tax_code, '') + ', ' + ISNULL(a.item_tax_code, '') + ', ' + ISNULL(c.code, '') + ', ' + ISNULL(CAST(a.price AS NVARCHAR), '') + ', ' + ISNULL(a.notes, '') + ', ' + ISNULL(e.long_description, '') AS edit_history
        FROM z_tbl_bpi_items_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_payment_terms c ON a.payment_terms_id = c.id
            LEFT JOIN tbl_setup_item d ON a.item_id = d.id
            LEFT JOIN tbl_setup_item_additional_specs e ON d.id = e.based_id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            ISNULL(b.first_name, '') + ',' + ISNULL(b.last_name, '') AS edit_by,
            ISNULL(CAST(a.date_added AS NVARCHAR), '') + ', ' + ISNULL(a.file_name, '') + ', ' + ISNULL(a.accreditation_added_by, '') AS edit_history
        FROM z_tbl_bpi_accreditation_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
    ) a
WHERE a.at_date BETWEEN DATEADD(DAY, -7, CAST(GETDATE() AS DATE))
    AND CAST(GETDATE() AS DATE)