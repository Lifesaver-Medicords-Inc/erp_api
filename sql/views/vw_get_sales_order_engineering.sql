-- Engineering's Sales Order screen (GET /engineering/job_order/sales_order).
--
-- Fixed 2026-09-05 (user report: "no customer displaying on engineering's Sales Order,
-- in normal sales order it was okay").
--
-- tbl_trans_sales_order.customer_id holds a tbl_bpi.id - the partner record - NOT a
-- tbl_bpi_general.id. This view joined it as though it were the latter:
--
--     LEFT JOIN tbl_bpi_general bpg ON so.customer_id = bpg.id     <- never matched
--     LEFT JOIN tbl_bpi         bpi ON bpg.based_id  = bpi.id      <- chained off it
--
-- so bpg never matched, customer came back NULL, and because the tbl_bpi join hung off
-- bpg, tin came back NULL too. Worked example: SO#0007 has customer_id 40034, which is
-- tbl_bpi.id 40034; its tbl_bpi_general row is id 40023 with based_id 40034. Joining on
-- based_id yields "Cavite power pumps" / "C#0016" / "777-777-777-77777" - exactly what
-- the sales app's own Sales Order screen shows for the same order.
--
-- `code` also pointed at so.doc, so the CODE box under CUSTOMER repeated the sales
-- order number ("0007") instead of the customer code ("C#0016"). doc_no already carries
-- so.doc for the DOC NO. field, so nothing loses it.
--
-- OUTER APPLY rather than a LEFT JOIN on based_id: a partner may have several
-- tbl_bpi_general rows (branches - two partners do today), and a plain join would
-- duplicate every sales order belonging to them. The sales order records only the
-- partner, not which branch, so there is nothing to disambiguate on - lowest id wins,
-- which is the original record rather than a later branch.
IF NOT EXISTS (SELECT 1 FROM sys.views WHERE name = 'vw_get_sales_order_engineering' AND schema_id = SCHEMA_ID('dbo'))
BEGIN
    EXEC('CREATE VIEW [dbo].[vw_get_sales_order_engineering] AS SELECT 1 AS placeholder')
END
GO
ALTER VIEW [dbo].[vw_get_sales_order_engineering] AS
SELECT so.order_id AS id,
    bpg.branch_name AS customer,
    bpi.tin AS tin,
    bpg.customer_code AS code,
    shipAddr.location AS delivery_to,
    billAddr.location AS bill_to,
    so.doc AS doc_no,
    so.date AS date,
    so.delivery_date AS delivery_date,
    so.ref_po AS reference_doc,
    so.status AS status
FROM dbo.tbl_trans_sales_order AS so
    LEFT JOIN dbo.tbl_bpi AS bpi ON bpi.id = so.customer_id
    OUTER APPLY (
        SELECT TOP 1 g.branch_name, g.customer_code
        FROM dbo.tbl_bpi_general AS g
        WHERE g.based_id = so.customer_id
        ORDER BY g.id
    ) bpg
    LEFT JOIN dbo.tbl_bpi_address AS shipAddr ON so.ship_to_id = shipAddr.id
    LEFT JOIN dbo.tbl_bpi_address AS billAddr ON so.bill_to_id = billAddr.id
