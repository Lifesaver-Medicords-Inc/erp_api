ALTER PROCEDURE [dbo].[sp_GetAPVoucherPayment] @SupplierId INT AS BEGIN
SET NOCOUNT ON;
BEGIN TRY
SELECT apv.id AS ap_voucher_id,
    apv.supplier,
    apv.supplier_code,
    apv.currency,
    apv.doc_no,
    apv.doc_date,
    -- Remaining Balance
    apv.transaction_amount - CASE
        WHEN ABS(ISNULL(SUM(pvd.amount_applied), 0)) < 0.01 THEN 0
        ELSE ISNULL(SUM(pvd.amount_applied), 0)
    END AS transaction_amount
FROM dbo.tbl_accounting_ap_voucher AS apv
    INNER JOIN dbo.tbl_accounting_ap_voucher_details apd ON apv.id = apd.ap_voucher_id
    LEFT JOIN dbo.tbl_accounting_payment_voucher_details pvd ON apd.id = pvd.ap_voucher_details_id
WHERE apv.supplier_id = @SupplierId
GROUP BY apv.id,
    apv.supplier,
    apv.supplier_code,
    apv.currency,
    apv.doc_no,
    apv.doc_date,
    apv.transaction_amount
HAVING apv.transaction_amount - CASE
        WHEN ABS(ISNULL(SUM(pvd.amount_applied), 0)) < 0.01 THEN 0
        ELSE ISNULL(SUM(pvd.amount_applied), 0)
    END > 0
END TRY BEGIN CATCH THROW;
END CATCH
END;