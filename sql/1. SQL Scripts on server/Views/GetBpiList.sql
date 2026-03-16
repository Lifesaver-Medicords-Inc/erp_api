CREATE VIEW [dbo].[GetBpiList]
AS
SELECT
    a.id,
    a.sales_id,
    a.name,
    a.main_website,
    a.tin,
    a.main_tel_no,
    c.industry_names,
    c.industry_ids
FROM tbl_bpi a
LEFT JOIN (
    SELECT 
        bi.bpi_id,

        -- Industry IDs
        STUFF((
            SELECT ',' + CAST(bi2.industry_id AS VARCHAR(50))
            FROM tbl_bpi_industries bi2
            WHERE bi2.bpi_id = bi.bpi_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS industry_ids,

        -- Industry Codes
        STUFF((
            SELECT ',' + si.code
            FROM tbl_bpi_industries bi3
            LEFT JOIN tbl_setup_bpi_industries si 
                ON bi3.industry_id = si.id
            WHERE bi3.bpi_id = bi.bpi_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS industry_names

    FROM tbl_bpi_industries bi
    GROUP BY bi.bpi_id
) c ON a.id = c.bpi_id
WHERE ISNULL(c.industry_names,'') <> '';
GO