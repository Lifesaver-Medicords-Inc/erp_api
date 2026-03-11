CREATE
OR ALTER VIEW [dbo].[vw_get_bpi_general_list] AS
SELECT a.id AS general_id,
    a.based_id AS general_based_id,
    a.social,
    a.branch_name,
    a.transaction_type,
    a.class_name,
    a.branch_tel_no,
    a.branch_website,
    a.customer_code,
    a.supplier_code,
    a.fax_no,
    a.notes,
    c.industry_ids AS branch_industry_ids,
    c.industry_names AS branch_industry_names,
    c.bpi_branch_id,
    d.entity_ids,
    d.entity_names,
    d.bpi_entity_id
FROM tbl_bpi_general a
    LEFT JOIN (
        SELECT ISNULL(STRING_AGG(b.industry_id, ','), '') AS industry_ids,
            ISNULL(STRING_AGG(b.code, ','), '') AS industry_names,
            b.bpi_general_id,
            a.id AS bpi_branch_id
        FROM tbl_bpi_branch_industries a
            LEFT JOIN (
                SELECT bb.id,
                    aa.bpi_general_id,
                    bb.code,
                    aa.industry_id
                FROM tbl_bpi_branch_industries aa
                    LEFT JOIN tbl_setup_bpi_industries bb ON aa.industry_id = bb.id
                GROUP BY bb.id,
                    aa.bpi_general_id,
                    bb.code,
                    aa.industry_id
            ) b ON a.bpi_general_id = b.bpi_general_id
            AND a.industry_id = b.id
        GROUP by b.bpi_general_id,
            a.id
    ) c ON a.id = c.bpi_general_id
    LEFT JOIN (
        SELECT ISNULL(STRING_AGG(b.entity_id, ','), '') AS entity_ids,
            ISNULL(STRING_AGG(b.code, ','), '') AS entity_names,
            b.bpi_general_id,
            a.id AS bpi_entity_id
        FROM tbl_bpi_entity a
            LEFT JOIN (
                SELECT bb.id,
                    aa.bpi_general_id,
                    bb.code,
                    aa.entity_id
                FROM tbl_bpi_entity aa
                    LEFT JOIN tbl_setup_bpi_entity bb ON aa.entity_id = bb.id
                GROUP BY bb.id,
                    aa.bpi_general_id,
                    bb.code,
                    aa.entity_id
            ) b ON a.bpi_general_id = b.bpi_general_id
            AND a.entity_id = b.id
        GROUP by b.bpi_general_id,
            a.id
    ) d ON a.id = d.bpi_general_id