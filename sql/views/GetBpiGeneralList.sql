ALTER VIEW [dbo].[GetBpiGeneralList] AS
SELECT a.id AS general_id,
    a.based_id AS general_based_id,
    a.social,
    a.branch_name,
    a.is_main,
    a.transaction_type,
    a.class_name,
    a.branch_tel_no,
    a.branch_website,
    a.customer_code,
    a.supplier_code,
    a.fax_no,
    a.notes,
    a.sales_id AS branch_sales_id,
    c.industry_ids AS branch_industry_ids,
    c.industry_names AS branch_industry_names,
    d.entity_ids,
    d.entity_names
FROM tbl_bpi_general a
    LEFT JOIN (
        SELECT bpi_general_id,
            STUFF(
                (
                    SELECT ',' + CAST(industry_id AS VARCHAR)
                    FROM tbl_bpi_branch_industries t2
                    WHERE t2.bpi_general_id = t1.bpi_general_id FOR XML PATH(''),
                        TYPE
                ).value('.', 'NVARCHAR(MAX)'),
                1,
                1,
                ''
            ) AS industry_ids,
            STUFF(
                (
                    SELECT ',' + bb.code
                    FROM tbl_bpi_branch_industries t2
                        LEFT JOIN tbl_setup_bpi_industries bb ON t2.industry_id = bb.id
                    WHERE t2.bpi_general_id = t1.bpi_general_id FOR XML PATH(''),
                        TYPE
                ).value('.', 'NVARCHAR(MAX)'),
                1,
                1,
                ''
            ) AS industry_names
        FROM tbl_bpi_branch_industries t1
        GROUP BY bpi_general_id
    ) c ON a.id = c.bpi_general_id
    LEFT JOIN (
        SELECT bpi_general_id,
            STUFF(
                (
                    SELECT ',' + CAST(entity_id AS VARCHAR)
                    FROM tbl_bpi_entity t2
                    WHERE t2.bpi_general_id = t1.bpi_general_id FOR XML PATH(''),
                        TYPE
                ).value('.', 'NVARCHAR(MAX)'),
                1,
                1,
                ''
            ) AS entity_ids,
            STUFF(
                (
                    SELECT ',' + bb.code
                    FROM tbl_bpi_entity t2
                        LEFT JOIN tbl_setup_bpi_entity bb ON t2.entity_id = bb.id
                    WHERE t2.bpi_general_id = t1.bpi_general_id FOR XML PATH(''),
                        TYPE
                ).value('.', 'NVARCHAR(MAX)'),
                1,
                1,
                ''
            ) AS entity_names
        FROM tbl_bpi_entity t1
        GROUP BY bpi_general_id
    ) d ON a.id = d.bpi_general_id