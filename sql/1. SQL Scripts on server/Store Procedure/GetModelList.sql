CREATE PROCEDURE [dbo].[GetModelList]
AS
BEGIN
    SELECT 
		a.*, 
		b.name AS related_name,  
		c.name AS related_brand
	FROM tbl_setup_item_model a
		LEFT JOIN tbl_setup_item_name b ON a.based_id = b.id
		LEFT JOIN tbl_setup_item_brand c ON a.based_id = c.id;
END;


GO
