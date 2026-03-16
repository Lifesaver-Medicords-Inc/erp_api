CREATE VIEW [dbo].[vw_items]
AS
SELECT 
    a.id,
    a.item_name_id,
    a.item_class_id,
    a.item_brand_id,
    a.unit_of_measure_id,
    a.item_tangibility_type,
    a.item_model,
    a.catalogue_year,
    a.price,
    a.item_code,
    a.is_stop_selling,
    b.name AS item_name,
    d.name AS item_class,
    e.name AS item_brand,
    f.name AS unit_of_measure
FROM tbl_setup_item a
LEFT JOIN tbl_setup_item_name b 
    ON a.item_name_id = b.id
LEFT JOIN tbl_setup_item_class d 
    ON a.item_class_id = d.id
LEFT JOIN tbl_setup_item_brand e 
    ON a.item_brand_id = e.id
LEFT JOIN tbl_setup_item_unit_measurement f 
    ON a.unit_of_measure_id = f.id


GO