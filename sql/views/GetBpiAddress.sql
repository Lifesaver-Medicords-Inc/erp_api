CREATE
OR ALTER VIEW [dbo].[GetBpiAddress] AS
SELECT a.id AS address_id,
    a.based_id as address_based_id,
    a.location
FROM tbl_bpi_address a