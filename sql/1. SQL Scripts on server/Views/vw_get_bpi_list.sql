CREATE VIEW [dbo].[vw_get_bpi_list] 
AS

SELECT  
    a.id AS id,  
    a.sales_id,  
    a.name AS name,  
    a.main_website,  
    a.tin,  
    a.main_tel_no,  
    c.industry_names,  
    c.industry_ids  

FROM tbl_bpi a  

LEFT JOIN
(
    SELECT
        x.bpi_id,

        ISNULL(
            STUFF((
                SELECT ',' + CAST(t.industry_id AS NVARCHAR(MAX))
                FROM tbl_bpi_industries t
                WHERE t.bpi_id = x.bpi_id
                FOR XML PATH(''), TYPE
            ).value('.', 'NVARCHAR(MAX)'),1,1,'')
        ,'') AS industry_ids,

        ISNULL(
            STUFF((
                SELECT ',' + s.code
                FROM tbl_bpi_industries t
                LEFT JOIN tbl_setup_bpi_industries s
                    ON t.industry_id = s.id
                WHERE t.bpi_id = x.bpi_id
                FOR XML PATH(''), TYPE
            ).value('.', 'NVARCHAR(MAX)'),1,1,'')
        ,'') AS industry_names

    FROM tbl_bpi_industries x
    GROUP BY x.bpi_id

) c ON a.id = c.bpi_id  

WHERE ISNULL(c.industry_names, '') <> ''

GO