CREATE PROCEDURE [dbo].[sp_GetBpiGeneralList]
AS
BEGIN
    SELECT
        a.id,
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

        -- Industry IDs
        ISNULL(STUFF((
            SELECT ',' + CAST(bi.industry_id AS VARCHAR(MAX))
            FROM tbl_bpi_branch_industries bi
            WHERE bi.bpi_general_id = a.id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,''),'') AS branch_industry_ids,

        -- Industry Names
        ISNULL(STUFF((
            SELECT ',' + si.code
            FROM tbl_bpi_branch_industries bi
            LEFT JOIN tbl_setup_bpi_industries si
                ON bi.industry_id = si.id
            WHERE bi.bpi_general_id = a.id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,''),'') AS branch_industry_names,

        -- Entity IDs
        ISNULL(STUFF((
            SELECT ',' + CAST(be.entity_id AS VARCHAR(MAX))
            FROM tbl_bpi_entity be
            WHERE be.bpi_general_id = a.id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,''),'') AS entity_ids,

        -- Entity Names
        ISNULL(STUFF((
            SELECT ',' + se.code
            FROM tbl_bpi_entity be
            LEFT JOIN tbl_setup_bpi_entity se
                ON be.entity_id = se.id
            WHERE be.bpi_general_id = a.id
            FOR XML PATH(''), TYPE
        ).value('.', 'NVARCHAR(MAX)'),1,1,''),'') AS entity_names

    FROM tbl_bpi_general a
END
GO