IF NOT EXISTS (SELECT 1 FROM sys.procedures WHERE name = 'sp_GetItemReleaseDetails' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE PROCEDURE [dbo].[sp_GetItemReleaseDetails] AS SET NOCOUNT ON;')
END
GO
ALTER PROCEDURE [dbo].[sp_GetItemReleaseDetails] @ItemReleaseId BIGINT AS BEGIN
SET NOCOUNT ON;
SELECT a.item_release_id,
    a.sales_order_details_id,
    a.item_id,
    a.released_qty,
    a.released_uom_id AS released_uom,
    c.item_code,
    a.serial_no
FROM tbl_inv_item_release_details a
    LEFT JOIN tbl_setup_item c ON a.item_id = c.id
WHERE a.item_release_id = @ItemReleaseId;
END