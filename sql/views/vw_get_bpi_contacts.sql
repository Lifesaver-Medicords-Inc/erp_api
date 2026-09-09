IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_contacts' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_contacts] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_contacts] AS
SELECT a.id AS contacts_id,
    a.number,
    a.based_id AS contacts_based_id,
    a.name,
    a.email,
    a.preferences,
    a.position,
    a.branch_id,
    a.is_default_contact,
    a.contact_notes AS contact_notes
FROM tbl_bpi_contacts a