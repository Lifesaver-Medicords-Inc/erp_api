ALTER PROCEDURE [dbo].[sp_GetBpiGeneralList] AS BEGIN
SELECT a.id,
    a.based_id,
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
    d.entity_ids,
    d.entity_names
FROM tbl_bpi_general a
    LEFT JOIN (
        SELECT ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(aa.industry_id AS NVARCHAR(MAX))
                        FROM tbl_bpi_branch_industries aa
                        WHERE aa.bpi_general_id = a.bpi_general_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS industry_ids,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + bb.code
                        FROM tbl_bpi_branch_industries aa
                            LEFT JOIN tbl_setup_bpi_industries bb ON aa.industry_id = bb.id
                        WHERE aa.bpi_general_id = a.bpi_general_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS industry_names,
            a.bpi_general_id
        FROM tbl_bpi_branch_industries a
        GROUP BY a.bpi_general_id
    ) c ON a.id = c.bpi_general_id
    LEFT JOIN (
        SELECT ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(aa.entity_id AS NVARCHAR(MAX))
                        FROM tbl_bpi_entity aa
                        WHERE aa.bpi_general_id = a.bpi_general_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS entity_ids,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + bb.code
                        FROM tbl_bpi_entity aa
                            LEFT JOIN tbl_setup_bpi_entity bb ON aa.entity_id = bb.id
                        WHERE aa.bpi_general_id = a.bpi_general_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS entity_names,
            a.bpi_general_id
        FROM tbl_bpi_entity a
        GROUP BY a.bpi_general_id
    ) d ON a.id = d.bpi_general_id
END;