ALTER VIEW [dbo].[GetBpiList] AS
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
        SELECT b.bpi_id,
            ISNULL(
                STUFF(
                    (
                        SELECT ',' + CAST(sub.industry_id AS VARCHAR)
                        FROM (
                                SELECT aa.bpi_id,
                                    aa.industry_id
                                FROM tbl_bpi_industries aa
                                WHERE aa.bpi_id = b.bpi_id
                                GROUP BY aa.bpi_id,
                                    aa.industry_id
                            ) sub FOR XML PATH(''),
                            TYPE
                    ).value('.', 'NVARCHAR(MAX)'),
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
                        FROM tbl_bpi_industries aa
                            LEFT JOIN tbl_setup_bpi_industries bb ON aa.industry_id = bb.id
                        WHERE aa.bpi_id = b.bpi_id
                        GROUP BY bb.code FOR XML PATH(''),
                            TYPE
                    ).value('.', 'NVARCHAR(MAX)'),
                    1,
                    1,
                    ''
                ),
                ''
            ) AS industry_names
        FROM (
                SELECT DISTINCT bpi_id
                FROM tbl_bpi_industries
            ) b
    ) c ON a.id = c.bpi_id
WHERE ISNULL(c.industry_names, '') <> ''