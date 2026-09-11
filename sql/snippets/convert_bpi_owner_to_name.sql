-- ============================================================
-- BPI ownership: employee id -> the owner's full NAME.
--
-- tbl_bpi.sales_id and tbl_bpi_general.sales_id used to hold the owner's
-- EMPLOYEE ID (IT-WD-1, Sales--13 ...). Management chose the NAME instead
-- ("Julie Cestona"), which is also what every other document already
-- stores: tbl_trans_sales_order.sales_executive, the Delivery Receipt's
-- sales_executive, and the Sales Invoice's sales_person all hold names.
-- The inventory and sales apps now compare ownership by name.
--
-- Full name = first_name + ' ' + last_name, trimmed - the same rule as
-- CurrentUserModel.full_name on the client, so the two always agree.
--
-- BACKUP FIRST: every row changed is recorded in
-- zz_bpi_owner_conversion_backup (table, row id, old value, new value).
-- Once an id becomes a name, reversing it depends on no user ever being
-- renamed - the backup removes that dependency. Drop it once you are
-- satisfied.
--
-- IDEMPOTENT: only values that ARE a user's employee id are converted, so
-- a second run changes nothing. Blank owners and the OFFICE house account
-- are left alone.
-- ============================================================
SET NOCOUNT ON;
GO

IF OBJECT_ID('dbo.zz_bpi_owner_conversion_backup') IS NULL
    CREATE TABLE dbo.zz_bpi_owner_conversion_backup (
        id            BIGINT IDENTITY(1,1) PRIMARY KEY,
        tbl           NVARCHAR(64)  NOT NULL,
        row_id        BIGINT        NOT NULL,
        old_sales_id  NVARCHAR(200) NULL,
        new_sales_id  NVARCHAR(200) NULL,
        converted_at  DATETIME      NOT NULL DEFAULT GETDATE()
    );
GO

INSERT INTO dbo.zz_bpi_owner_conversion_backup (tbl, row_id, old_sales_id, new_sales_id)
SELECT 'tbl_bpi', b.id, b.sales_id,
       LTRIM(RTRIM(ISNULL(LTRIM(RTRIM(u.first_name)), '') + ' ' + ISNULL(LTRIM(RTRIM(u.last_name)), '')))
FROM tbl_bpi b JOIN tbl_setup_users u ON u.employee_id = b.sales_id
WHERE ISNULL(b.sales_id, '') <> '';

INSERT INTO dbo.zz_bpi_owner_conversion_backup (tbl, row_id, old_sales_id, new_sales_id)
SELECT 'tbl_bpi_general', g.id, g.sales_id,
       LTRIM(RTRIM(ISNULL(LTRIM(RTRIM(u.first_name)), '') + ' ' + ISNULL(LTRIM(RTRIM(u.last_name)), '')))
FROM tbl_bpi_general g JOIN tbl_setup_users u ON u.employee_id = g.sales_id
WHERE ISNULL(g.sales_id, '') <> '';

DECLARE @b INT, @g INT;

UPDATE b SET b.sales_id = LTRIM(RTRIM(ISNULL(LTRIM(RTRIM(u.first_name)), '') + ' ' + ISNULL(LTRIM(RTRIM(u.last_name)), '')))
FROM tbl_bpi b JOIN tbl_setup_users u ON u.employee_id = b.sales_id
WHERE ISNULL(b.sales_id, '') <> '';
SET @b = @@ROWCOUNT;

UPDATE g SET g.sales_id = LTRIM(RTRIM(ISNULL(LTRIM(RTRIM(u.first_name)), '') + ' ' + ISNULL(LTRIM(RTRIM(u.last_name)), '')))
FROM tbl_bpi_general g JOIN tbl_setup_users u ON u.employee_id = g.sales_id
WHERE ISNULL(g.sales_id, '') <> '';
SET @g = @@ROWCOUNT;

PRINT N'companies converted : ' + CAST(@b AS NVARCHAR(20));
PRINT N'branches converted  : ' + CAST(@g AS NVARCHAR(20));
GO
