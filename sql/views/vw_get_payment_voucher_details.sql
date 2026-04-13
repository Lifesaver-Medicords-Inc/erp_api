ALTER VIEW [dbo].[vw_get_payment_voucher_details] AS
SELECT pvd.id,
    pvd.payment_voucher_id,
    pvd.ap_voucher_details_id,
    pvd.doc_no,
    pvd.due_date,
    pvd.trans_amount,
    pvd.amount_applied,
    pvd.twas_applied,
    -- Total applied per AP voucher detail
    ISNULL(
        SUM(pvd.amount_applied) OVER (PARTITION BY pvd.ap_voucher_details_id),
        0
    ) AS total_amount_applied,
    -- Open amount = original line_amount - total applied
    apd.line_amount - ISNULL(
        SUM(pvd.amount_applied) OVER (PARTITION BY pvd.ap_voucher_details_id),
        0
    ) AS open_amount
FROM tbl_accounting_payment_voucher_details pvd
    INNER JOIN tbl_accounting_ap_voucher_details apd ON pvd.ap_voucher_details_id = apd.id;