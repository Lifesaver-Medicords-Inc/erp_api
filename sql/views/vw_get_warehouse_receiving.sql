IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_warehouse_receiving' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_warehouse_receiving] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_warehouse_receiving] AS
SELECT wn.id AS warehouse_id,
    wn.name AS warehouse,
    CONCAT(
        wa.building_no,
        ', ',
        wa.street,
        ', ',
        wa.barangay_no,
        ', ',
        wa.city
    ) AS warehouse_address
FROM tbl_inv_warehouse_name wn
    INNER JOIN tbl_inv_warehouse_address wa ON wn.id = wa.warehouse_name_id
WHERE is_inactive = 0;