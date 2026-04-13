ALTER VIEW [dbo].[vw_get_active_warehouse] AS
SELECT *
FROM tbl_inv_warehouse_name
WHERE is_inactive = 0;