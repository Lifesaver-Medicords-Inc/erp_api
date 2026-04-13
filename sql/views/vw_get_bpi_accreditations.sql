ALTER VIEW [dbo].[vw_get_bpi_accreditations] AS
SELECT a.id AS bpi_accreditation_id,
    a.based_id AS bpi_accreditation_based_id,
    a.branch_id AS bpi_accreditation_branch_id,
    a.date_added,
    a.file_name,
    a.accreditation_added_by,
    a.file_path
FROM tbl_bpi_accreditation a