-- Creates z_tbl_setup_item_at, the Item Entry history table, where it is missing.
--
-- UpdateItem writes a history row on every item save (services/setup_services/item_service.go).
-- Lightspeed_test_fresh never had this table - no startup migration creates it - so every item
-- update against that database failed with "failed inserting itemat" and was rolled back
-- (found 2026-09-14). The definition matches Lightspeed_ERP_deploy_rehearsal's copy exactly.
-- Safe to run more than once.

IF OBJECT_ID('dbo.z_tbl_setup_item_at', 'U') IS NULL
BEGIN
    CREATE TABLE dbo.z_tbl_setup_item_at (
        id                    bigint IDENTITY(1,1) NOT NULL PRIMARY KEY CLUSTERED,
        ref_id                bigint NULL,
        item_name_id          bigint NULL,
        item_class_id         bigint NULL,
        item_brand_id         bigint NULL,
        unit_of_measure_id    bigint NULL,
        item_model            nvarchar(max) NULL,
        catalogue_year        nvarchar(max) NULL,
        item_code             nvarchar(max) NULL,
        item_tangibility_type nvarchar(max) NULL,
        is_stop_selling       bit NULL,
        price                 float NULL,
        AT_ACTION             nvarchar(max) NULL,
        IP_ADDRESS            nvarchar(max) NULL,
        MOTHERBOARD_SERIAL_NO nvarchar(max) NULL,
        MACHINE_NAME          nvarchar(max) NULL,
        AT_DATE               nvarchar(max) NULL,
        AT_USER_ID            nvarchar(max) NULL,
        AT_USER               nvarchar(max) NULL
    );
END
