CREATE
OR ALTER VIEW [dbo].[get_customer] AS
SELECT *
FROM (
        SELECT a.id AS based_id,
            a.id AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(
                ',',
                a.name,
                a.main_tel_no,
                a.main_website,
                a.tin
            ) AS edit_history
        FROM z_tbl_bpi_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
        UNION ALL
        SELECT a.id,
            '' AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(
                ', ',
                a.branch_name,
                a.transaction_type,
                a.class_name,
                a.branch_tel_no,
                a.branch_website,
                a.customer_code,
                c.code,
                a.supplier_code,
                a.fax_no,
                a.notes
            ) AS edit_history
        FROM z_tbl_bpi_general_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_social c ON a.social = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(
                ', ',
                a.name,
                a.number,
                a.name,
                a.email,
                a.preferences,
                c.code
            ) AS edit_history
        FROM z_tbl_bpi_contacts_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_social c ON a.position = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS('', a.location, ' ') AS edit_history
        FROM z_tbl_bpi_address_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
        UNION ALL
        SELECT a.finance_based_id AS based_id,
            a.finance_branch_id AS branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(', ', a.finance_account_id, c.code) AS edit_history
        FROM z_tbl_bpi_finance_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
            LEFT JOIN tbl_setup_payment_terms c ON a.finance_payment_terms_id = c.id
        UNION ALL
        SELECT a.based_id,
            a.branch_id,
            a.AT_DATE AS at_date,
            a.AT_ACTION AS actions,
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(
                ', ',
                a.tax_code,
                a.item_tax_code,
                c.code,
                a.price,
                a.notes,
                e.long_description
            ) AS edit_history
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
            CONCAT_WS(',', b.first_name, b.last_name) AS edit_by,
            CONCAT_WS(
                ', ',
                a.date_added,
                a.file_name,
                a.accreditation_added_by
            ) AS edit_history
        FROM z_tbl_bpi_accreditation_at a
            LEFT JOIN tbl_setup_users b ON a.AT_USER_ID = b.id
    ) a
WHERE a.at_date BETWEEN DATEADD(DAY, -7, CAST(GETDATE() AS DATE))
    AND CAST(GETDATE() AS DATE)