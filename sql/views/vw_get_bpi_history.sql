IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bpi_history' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bpi_history] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bpi_history] AS
SELECT based_id AS branch_id,
    date,
    actions,
    CONCAT(
        CASE
            WHEN actions = 'update' THEN 'Updated '
            WHEN actions = 'create' THEN 'Created '
            WHEN actions = 'delete' THEN 'Deleted '
            ELSE ''
        END,
        child_type
    ) AS edit_history,
    edit_by
FROM tbl_bpi_history;