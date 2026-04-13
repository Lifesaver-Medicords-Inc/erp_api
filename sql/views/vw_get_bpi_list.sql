ALTER VIEW [dbo].[vw_get_bpi_list] AS
SELECT a.id AS id,
    a.sales_id,
    a.name AS name,
    a.main_website,
    a.tin,
    a.main_tel_no,
    c.industry_names,
    c.industry_ids
FROM tbl_bpi a
    LEFT JOIN (
        SELECT a.bpi_id,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(aa.industry_id AS NVARCHAR(MAX))
                        FROM tbl_bpi_industries aa
                        WHERE aa.bpi_id = a.bpi_id FOR XML PATH('')
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
                        SELECT ',' + CAST(bb.code AS NVARCHAR(MAX))
                        FROM tbl_bpi_industries aa
                            LEFT JOIN tbl_setup_bpi_industries bb ON aa.industry_id = bb.id
                        WHERE aa.bpi_id = a.bpi_id FOR XML PATH('')
                    ),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS industry_names
        FROM tbl_bpi_industries a
        GROUP BY a.bpi_id
    ) c ON a.id = c.bpi_id
WHERE ISNULL(c.industry_names, '') <> '';