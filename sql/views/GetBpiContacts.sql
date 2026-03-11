CREATE
OR ALTER VIEW [dbo].[GetBpiContacts] AS
SELECT a.id AS contacts_id,
    a.number,
    a.based_id AS contacts_based_id,
    a.name,
    a.email,
    a.preferences,
    a.position
FROM tbl_bpi_contacts a