ALTER VIEW [dbo].[GetWiringNotes] AS
SELECT a.id as wiring_id,
    a.based_id,
    a.materials,
    NULL AS item_description,
    a.num_of_wires_set,
    a.num_of_qty_set,
    a.distance_travelled_set,
    a.allowance_wire_set,
    a.num_of_sets,
    a.total_qty,
    d.note_id,
    d.wiring_note
FROM tbl_trans_sales_project_wiring a
    LEFT JOIN tbl_trans_sales_project_item_set b ON a.based_id = b.item_set_id
    LEFT JOIN tbl_wiring_user_inputs d ON a.id = d.wiring_id