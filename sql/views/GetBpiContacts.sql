IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'GetBpiContacts' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[GetBpiContacts] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[GetBpiContacts] AS
SELECT a.id AS contacts_id,
    a.number,
    a.based_id AS contacts_based_id,
    a.name,
    a.email,
    a.preferences,
    a.position
FROM tbl_bpi_contacts a