IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_bin_loc_pick_activity' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_bin_loc_pick_activity] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_bin_loc_pick_activity] AS
SELECT a.warehouse_name_id,
    a.zone,
    a.area,
    a.rack,
    a.level,
    a.bins,
    a.location_code,
    b.name AS warehouse_name
FROM tbl_inv_warehouse_area a
    INNER JOIN tbl_inv_warehouse_name b ON a.warehouse_name_id = b.id