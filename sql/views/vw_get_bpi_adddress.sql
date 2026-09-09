IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_adddress' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_adddress] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_adddress] AS
SELECT a.id AS address_ids,
    a.based_id as address_based_id,
    a.location,
    a.branch_id as address_branch_id,
    a.is_deleted as address_is_deleted
FROM tbl_bpi_address a