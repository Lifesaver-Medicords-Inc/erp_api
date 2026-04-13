ALTER VIEW [dbo].[vw_get_entity_count] AS
SELECT c.code,
    COUNT(c.code) AS entity_count
FROM tbl_bpi_entity a
    LEFT JOIN tbl_setup_bpi_entity c ON a.entity_id = c.id
GROUP BY c.code