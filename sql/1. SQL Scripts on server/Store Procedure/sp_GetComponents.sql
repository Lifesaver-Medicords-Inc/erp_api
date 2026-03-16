CREATE PROCEDURE [dbo].[sp_GetComponents]
    @BomId int
AS
BEGIN
    SET NOCOUNT ON;

    BEGIN TRY
        SELECT
            bod.id AS id,
            itn.name AS name,
			bod.bom_qty AS quantity,
			bod.bom_qty AS stock
        FROM dbo.tbl_trans_sales_order_details AS so
		INNER JOIN dbo.tbl_setup_item_bom AS bom
			ON so.bom_id = bom.id
		INNER JOIN dbo.tbl_setup_item_bom_details AS bod
			ON bom.id = bod.item_bom_id
		INNER JOIN dbo.tbl_setup_item AS i
			ON bod.item_id = i.id
		INNER JOIN dbo.tbl_setup_item_name AS itn
			ON i.item_name_id = itn.id
        WHERE so.bom_id = @BomId;
    END TRY
    BEGIN CATCH
        THROW;
    END CATCH
END;


GO
