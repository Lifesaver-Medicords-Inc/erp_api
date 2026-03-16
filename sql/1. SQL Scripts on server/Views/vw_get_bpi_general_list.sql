CREATE VIEW [dbo].[vw_get_bpi_general_list] 
AS
SELECT
    a.id AS general_id,
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

LEFT JOIN
(
    SELECT
        x.bpi_general_id,

        STUFF((
            SELECT ',' + CAST(t.industry_id AS VARCHAR(50))
            FROM tbl_bpi_branch_industries t
            WHERE t.bpi_general_id = x.bpi_general_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS industry_ids,

        STUFF((
            SELECT ',' + s.code
            FROM tbl_bpi_branch_industries t
            LEFT JOIN tbl_setup_bpi_industries s 
                ON t.industry_id = s.id
            WHERE t.bpi_general_id = x.bpi_general_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS industry_names,

        MIN(x.id) AS bpi_branch_id

    FROM tbl_bpi_branch_industries x
    GROUP BY x.bpi_general_id

) c ON a.id = c.bpi_general_id


LEFT JOIN
(
    SELECT
        x.bpi_general_id,

        STUFF((
            SELECT ',' + CAST(t.entity_id AS VARCHAR(50))
            FROM tbl_bpi_entity t
            WHERE t.bpi_general_id = x.bpi_general_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS entity_ids,

        STUFF((
            SELECT ',' + s.code
            FROM tbl_bpi_entity t
            LEFT JOIN tbl_setup_bpi_entity s 
                ON t.entity_id = s.id
            WHERE t.bpi_general_id = x.bpi_general_id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,'') AS entity_names,

        MIN(x.id) AS bpi_entity_id

    FROM tbl_bpi_entity x
    GROUP BY x.bpi_general_id

) d ON a.id = d.bpi_general_id



GO


