-- Grant the two Sales Order approval codes (spec §3.3) to the Positions allowed to use
-- them: Sales Manager and the Chief Business Development Officer, plus Admin (§3.2 gives
-- Admin every module).
--
-- WHY THIS IS NEEDED, AND WHY IT IS NOT AUTOMATIC
-- 'Sales - Order.Orders.Approve' and 'Sales - Order.Orders.Cancel Order' are new codes.
-- SeedAccessModules adds them to the tbl_access_modules CATALOG on startup, which is what
-- makes them appear in the Access Control tree - but seeding a catalog entry is not a
-- grant. Until some Position is actually granted them, the ✓ and CANCEL buttons are
-- hidden for everyone and the API refuses the status change, so no Sales Order can be
-- approved at all. Run this once per database after the API has started (the catalog rows
-- must exist first).
--
-- Idempotent: re-running it grants nothing twice. Matching Positions by NAME rather than a
-- hardcoded id, since ids differ between databases.
--
-- To grant to more Positions later, use the Access Control screen rather than editing this
-- file - it exists only to bootstrap the two Positions the spec names.

SET NOCOUNT ON;

DECLARE @codes TABLE (code NVARCHAR(255));
INSERT INTO @codes (code) VALUES
    (N'Sales - Order.Orders.Approve'),
    (N'Sales - Order.Orders.Cancel Order');

DECLARE @positions TABLE (name NVARCHAR(255));
INSERT INTO @positions (name) VALUES
    (N'Sales Manager'),
    (N'Chief Business Development Officer'),
    (N'Admin');

INSERT INTO tbl_position_access (position_id, code)
SELECT p.id, c.code
FROM tbl_position p
CROSS JOIN @codes c
WHERE p.name IN (SELECT name FROM @positions)
  AND NOT EXISTS (
      SELECT 1 FROM tbl_position_access pa
      WHERE pa.position_id = p.id AND pa.code = c.code
  );

PRINT CONCAT('Granted ', @@ROWCOUNT, ' new Sales Order approval permission(s).');

-- Verify
SELECT p.name AS position_name, pa.code
FROM tbl_position_access pa
INNER JOIN tbl_position p ON p.id = pa.position_id
WHERE pa.code IN (SELECT code FROM @codes)
ORDER BY p.name, pa.code;
