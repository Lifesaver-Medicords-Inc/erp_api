IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'GetBpiAddress' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[GetBpiAddress] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[GetBpiAddress] AS
SELECT a.id AS address_id,
    a.based_id as address_based_id,
    a.location
FROM tbl_bpi_address a