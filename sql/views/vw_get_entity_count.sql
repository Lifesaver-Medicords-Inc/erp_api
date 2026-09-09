IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_entity_count' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_entity_count] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_entity_count] AS
SELECT c.code,
    COUNT(c.code) AS entity_count
FROM tbl_bpi_entity a
    LEFT JOIN tbl_setup_bpi_entity c ON a.entity_id = c.id
GROUP BY c.code