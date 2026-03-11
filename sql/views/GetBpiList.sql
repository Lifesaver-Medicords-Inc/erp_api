CREATE
OR ALTER VIEW [dbo].[GetBpiList] AS
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
        SELECT ISNULL(STRING_AGG(b.industry_id, ','), '') AS industry_ids,
            ISNULL(STRING_AGG(b.code, ','), '') AS industry_names,
            b.bpi_id
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
        GROUP by b.bpi_id
    ) c on a.id = c.bpi_id
WHERE ISNULL(c.industry_names, '') <> ''