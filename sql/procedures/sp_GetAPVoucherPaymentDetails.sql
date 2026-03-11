CREATE
OR ALTER PROCEDURE [dbo].[sp_GetAPVoucherPaymentDetails] @VoucherId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT apd.id AS ap_voucher_details_id,
    apd.ap_voucher_id,
    apd.receipt_no AS doc_no,
    apd.ir_due_date AS due_date,
    apd.line_amount AS trans_amount,
    apd.twas_amount AS twas_applied,
    -- Sum of applied amounts (default to 0 if NULL)
    ISNULL(SUM(CAST(pvd.amount_applied AS FLOAT)), 0) AS total_applied,
    -- Compute open amount
    apd.line_amount - ISNULL(SUM(CAST(pvd.amount_applied AS FLOAT)), 0) AS open_amount
FROM dbo.tbl_accounting_ap_voucher_details AS apd
    LEFT JOIN dbo.tbl_accounting_payment_voucher_details AS pvd ON apd.id = pvd.ap_voucher_details_id
WHERE apd.ap_voucher_id = @VoucherId
GROUP BY apd.id,
    apd.ap_voucher_id,
    apd.receipt_no,
    apd.ir_due_date,
    apd.line_amount,
    apd.twas_amount -- Only include rows where open_amount > 0
HAVING (
        apd.line_amount - ISNULL(SUM(CAST(pvd.amount_applied AS FLOAT)), 0)
    ) > 0
END TRY BEGIN CATCH THROW;
END CATCH
END;