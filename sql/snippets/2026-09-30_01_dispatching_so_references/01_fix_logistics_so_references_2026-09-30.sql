-- STEP 01 of 01 - logistics schedules and routes: put the sales order's own number back
-- Safe to run on any database (rehearsal, test_fresh, production); run it once per database.
-- A TRIAL RUN until you set @commit = 1; it must end by printing "committed".
--
-- What happened
-- =============
-- Every delivery receipt creates (or reuses) a logistics schedule for its sales order and
-- adds a route to it. Both carry the order as a reference - the schedule in
-- sales_order_doc_no, the route in reference_doc - and the API filled it from
-- vw_get_sales_order_dr.document_no. On a sales order that column is NOT the order's number:
-- it holds the code of the quotation the order was made from ("0020" before quotation
-- versioning, "FQ#2026-0020.1.0" since). So the schedule and route of SO#0005 were labelled
-- 0020, SO#0020's number. The dispatching Sales Order screen matches schedules on the
-- reference's digits as well as on sales_order_id, so once SO#0020 exists it would also pick
-- up SO#0005's deliveries. The API now writes the order's own doc (fixed 2026-09-30,
-- delivery_receipt_service.go getSalesOrderDocNo); this puts the rows already written right.
--
-- What it changes, in one transaction
-- ===================================
--   * a schedule tied to an order (sales_order_id) whose sales_order_doc_no is not that
--     order's doc: set to the order's doc;
--   * a route made for a delivery receipt (delivery_receipt_doc = the DR's number, one DR with
--     that number) whose reference_doc is not the DR's order: set to that order's doc.
-- The number is stored bare ("0005"), as the API writes it; the screens show it as SO#0005.
-- Leaves alone: schedules with no sales_order_id, routes with no delivery receipt (their
-- reference is whatever the dispatcher chose), and references that already name the right
-- order either as "0005" or "SO#0005".
--
-- HOW TO RUN: sqlcmd -S <server> -d <database> -b -I -i <this file>
-- (or SSMS - CHECK THE DATABASE DROPDOWN). After a commit, restart that database's API or
-- clear its Redis keys, so the calendar does not keep serving the old labels from cache.
SET NOCOUNT ON;
SET XACT_ABORT ON;

DECLARE @commit bit = 0;   -- 1 = keep the changes, 0 = trial run

PRINT 'Database: ' + DB_NAME();

-- ---------------------------------------------------------------- what would change
IF OBJECT_ID('tempdb..#schedules') IS NOT NULL DROP TABLE #schedules;
IF OBJECT_ID('tempdb..#routes') IS NOT NULL DROP TABLE #routes;

SELECT s.id,
       s.sales_order_id,
       s.sales_order_doc_no AS old_ref,
       o.doc AS new_ref
  INTO #schedules
  FROM tbl_dispatching_logistics_calendar_schedule s
  JOIN tbl_trans_sales_order o ON o.order_id = s.sales_order_id
 WHERE s.sales_order_id > 0
   AND ISNULL(o.doc, '') <> ''
   AND ISNULL(s.sales_order_doc_no, '') NOT IN (o.doc, 'SO#' + o.doc);

SELECT r.id,
       r.schedule_id,
       r.delivery_receipt_doc,
       d.sales_order_id,
       r.reference_doc AS old_ref,
       o.doc AS new_ref
  INTO #routes
  FROM tbl_dispatching_logistics_route r
  JOIN (SELECT CAST(doc_no AS nvarchar(20)) AS doc_no, MIN(sales_order_id) AS sales_order_id
          FROM tbl_dispatching_delivery_receipt
         GROUP BY doc_no
        HAVING COUNT(*) = 1) d ON d.doc_no = r.delivery_receipt_doc
  JOIN tbl_trans_sales_order o ON o.order_id = d.sales_order_id
 WHERE ISNULL(o.doc, '') <> ''
   AND ISNULL(r.reference_doc, '') NOT IN (o.doc, 'SO#' + o.doc);

PRINT 'Schedules to correct:';
SELECT id AS schedule_id, sales_order_id, old_ref, new_ref FROM #schedules ORDER BY id;
PRINT 'Routes to correct:';
SELECT id AS route_id, schedule_id, delivery_receipt_doc AS dr_no, sales_order_id, old_ref, new_ref FROM #routes ORDER BY id;

-- ---------------------------------------------------------------- the change
BEGIN TRANSACTION;

UPDATE s
   SET s.sales_order_doc_no = c.new_ref
  FROM tbl_dispatching_logistics_calendar_schedule s
  JOIN #schedules c ON c.id = s.id;
PRINT CAST(@@ROWCOUNT AS varchar(10)) + ' schedule(s) updated';

UPDATE r
   SET r.reference_doc = c.new_ref
  FROM tbl_dispatching_logistics_route r
  JOIN #routes c ON c.id = r.id;
PRINT CAST(@@ROWCOUNT AS varchar(10)) + ' route(s) updated';

-- Nothing the rules above cover may still be wrong.
DECLARE @left int =
    (SELECT COUNT(*)
       FROM tbl_dispatching_logistics_calendar_schedule s
       JOIN tbl_trans_sales_order o ON o.order_id = s.sales_order_id
      WHERE s.sales_order_id > 0 AND ISNULL(o.doc, '') <> ''
        AND ISNULL(s.sales_order_doc_no, '') NOT IN (o.doc, 'SO#' + o.doc))
  + (SELECT COUNT(*)
       FROM tbl_dispatching_logistics_route r
       JOIN (SELECT CAST(doc_no AS nvarchar(20)) AS doc_no, MIN(sales_order_id) AS sales_order_id
               FROM tbl_dispatching_delivery_receipt
              GROUP BY doc_no
             HAVING COUNT(*) = 1) d ON d.doc_no = r.delivery_receipt_doc
       JOIN tbl_trans_sales_order o ON o.order_id = d.sales_order_id
      WHERE ISNULL(o.doc, '') <> ''
        AND ISNULL(r.reference_doc, '') NOT IN (o.doc, 'SO#' + o.doc));

IF @left <> 0
BEGIN
    ROLLBACK TRANSACTION;
    RAISERROR('%d reference(s) still wrong after the update - rolled back, nothing changed.', 16, 1, @left);
    RETURN;
END

IF @commit = 1
BEGIN
    COMMIT TRANSACTION;
    PRINT 'committed';
END
ELSE
BEGIN
    ROLLBACK TRANSACTION;
    PRINT 'trial run - rolled back, nothing changed (set @commit = 1 to keep it)';
END
