CREATE VIEW [dbo].[vw_get_bpi_history] AS
SELECT
    based_id AS branch_id,
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


GO
