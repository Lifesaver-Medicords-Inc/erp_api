CREATE
OR ALTER VIEW [dbo].[vw_get_bpi_list] AS
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
                STRING_AGG(CAST(b.industry_id AS NVARCHAR(MAX)), ','),
                ''
            ) AS industry_ids,
            ISNULL(
                STRING_AGG(CAST(b.code AS NVARCHAR(MAX)), ','),
                ''
            ) AS industry_names
        FROM tbl_bpi_industries a
            LEFT JOIN (
                SELECT bb.id,
                    aa.bpi_id,
                    bb.code,
                    aa.industry_id
                FROM tbl_bpi_industries aa
                    LEFT JOIN tbl_setup_bpi_industries bb ON aa.industry_id = bb.id
                GROUP BY bb.id,
                    aa.bpi_id,
                    bb.code,
                    aa.industry_id
            ) b ON a.bpi_id = b.bpi_id
            AND a.industry_id = b.id
        GROUP BY b.bpi_id
    ) c ON a.id = c.bpi_id