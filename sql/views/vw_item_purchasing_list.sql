CREATE
OR ALTER VIEW [dbo].[vw_item_purchasing_list] AS
SELECT a.id,
    a.item_id as based_id,
    a.based_id as supplier_name_id,
    a.payment_terms_id,
    a.price,
    b.name as supplier_name,
    c.name as payment_terms_name,
    d.class_name as supplier_type_name
FROM tbl_bpi_items a
    LEFT JOIN tbl_bpi b ON a.based_id = b.id
    LEFT JOIN tbl_setup_payment_terms c ON a.payment_terms_id = c.id
    LEFT JOIn tbl_bpi_general d ON b.id = d.based_id