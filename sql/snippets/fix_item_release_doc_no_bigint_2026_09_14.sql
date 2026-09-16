-- Makes tbl_inv_item_release.doc_no a bigint wherever it is still text.
--
-- Lightspeed_ERP_deploy_rehearsal had doc_no as nvarchar(50) with the unique constraint
-- uni_tbl_inv_item_release_doc_no, left behind by a `unique` tag removed from the model in 7734267
-- (2026-02-28). The model says int, so AutoMigrate tried to ALTER the column on every start and the
-- constraint blocked it - for ItemRelease and, through its foreign key, ItemReleaseDetails.
-- utils.NextDocNo also takes MAX(doc_no), and on text '9' sorts after '10', so the eleventh release
-- would have drawn 10 again and been rejected.
--
-- The constraint is dropped, not recreated: the model no longer declares it, GORM drops it itself
-- once the column type matches, and Lightspeed_test_fresh never had it. NextDocNo already reads
-- WITH (UPDLOCK, HOLDLOCK), so two releases cannot draw the same number.
--
-- Refuses to run if any doc_no is not a whole number. Safe to run more than once. (Found 2026-09-14.)

SET XACT_ABORT ON;

IF EXISTS (SELECT 1 FROM sys.columns c JOIN sys.types t ON t.user_type_id = c.user_type_id
           WHERE c.object_id = OBJECT_ID('dbo.tbl_inv_item_release') AND c.name = 'doc_no' AND t.name <> 'bigint')
BEGIN
    IF EXISTS (SELECT 1 FROM dbo.tbl_inv_item_release
               WHERE doc_no IS NOT NULL AND (TRY_CONVERT(bigint, doc_no) IS NULL OR LTRIM(RTRIM(doc_no)) = ''))
    BEGIN
        RAISERROR('tbl_inv_item_release.doc_no holds values that are not whole numbers; fix them first.', 16, 1);
        RETURN;
    END

    BEGIN TRANSACTION;

    IF EXISTS (SELECT 1 FROM sys.key_constraints
               WHERE name = 'uni_tbl_inv_item_release_doc_no' AND parent_object_id = OBJECT_ID('dbo.tbl_inv_item_release'))
        ALTER TABLE dbo.tbl_inv_item_release DROP CONSTRAINT uni_tbl_inv_item_release_doc_no;

    ALTER TABLE dbo.tbl_inv_item_release ALTER COLUMN doc_no bigint NULL;

    COMMIT TRANSACTION;
END
