-- Central status-derivation engine for one SO line item (spec §7.1). "The status is a
-- pure function of which document last referenced that item" - this checks the LATEST
-- possible stage first (delivery outcome) and falls back through earlier stages,
-- ending at the base stock-availability check, so whichever document is furthest along
-- the pipeline wins.
--
-- Column/table names verified directly against the live DB before writing this -
-- notably compatibility_level 110 (SQL Server 2012) rules out STRING_SPLIT, hence the
-- LIKE-based comma match against tbl_purchasing_purchase_order_details.order_detail_ids
-- rather than a cleaner set-based split.
--
-- Called by one trigger per source table (see sql/triggers/tr_so_item_status_*.sql),
-- each passing whichever order_details_id(s) their own change affects. Never called
-- directly by application code - CLAUDE.md's own accounting-inverts-spec-wins aside,
-- this deliberately replaces sp_SetOrderStatus's narrower allocated_qty-only logic
-- (SO_Item_Status_Module_Spec_2026-08-13.md §5 point 4) as the one place status gets
-- written, so nothing can update stock/PO/production/pick/release/DR and forget to
-- recompute it.
CREATE OR ALTER PROCEDURE [dbo].[sp_RecomputeSoItemStatus]
    @order_details_id BIGINT
AS
BEGIN
    SET NOCOUNT ON;

    DECLARE @item_id BIGINT, @required_qty BIGINT, @so_id BIGINT, @quotation_id BIGINT;
    DECLARE @new_status NVARCHAR(50) = NULL;

    SELECT @item_id = sod.item_id, @required_qty = sod.qty, @so_id = sod.based_id
    FROM tbl_trans_sales_order_details sod
    WHERE sod.order_details_id = @order_details_id;

    IF @item_id IS NULL
        RETURN; -- line no longer exists (deleted or id never valid) - nothing to compute

    SELECT @quotation_id = so.quotation_id
    FROM tbl_trans_sales_order so
    WHERE so.order_id = @so_id;

    -- 1. Delivery outcome (§7.1 rows 15-17) - latest possible stage.
    DECLARE @dr_doc_no BIGINT, @departed DATETIMEOFFSET, @arrived DATETIMEOFFSET, @returned DATETIMEOFFSET;

    SELECT TOP 1 @dr_doc_no = dr.doc_no
    FROM tbl_dispatching_delivery_receipt_items dri
    INNER JOIN tbl_dispatching_delivery_receipt dr ON dr.id = dri.delivery_receipt_id
    WHERE dri.sales_order_details_id = @order_details_id
    ORDER BY dr.id DESC;

    IF @dr_doc_no IS NOT NULL
    BEGIN
        -- LogisticsRoute links back to a DR only by doc-number string
        -- (delivery_receipt_doc), not a numeric FK - doc_no is uniquely indexed so this
        -- is reliable, just not as clean as an id join would be.
        SELECT TOP 1 @departed = lr.departed_at, @arrived = lr.arrived_at, @returned = lr.returned_at
        FROM tbl_dispatching_logistics_route lr
        WHERE lr.delivery_receipt_doc = CAST(@dr_doc_no AS NVARCHAR(50))
        ORDER BY lr.id DESC;

        IF @arrived IS NOT NULL
            SET @new_status = 'DELIVERED';
        ELSE IF @departed IS NOT NULL AND @returned IS NOT NULL
            SET @new_status = 'FAILED/RETURN';
        ELSE
            SET @new_status = 'FOR DELIVERY';
    END

    -- 2. Item Release (§7.1 rows 13-14) - "first half/second half" is quantity-based
    -- (required_qty vs released_qty), not a literal two-step workflow - confirmed from
    -- the schema itself, not assumed.
    IF @new_status IS NULL
    BEGIN
        DECLARE @rel_required BIGINT, @rel_released BIGINT;

        SELECT @rel_required = SUM(ird.required_qty), @rel_released = SUM(ISNULL(ird.released_qty, 0))
        FROM tbl_inv_item_release_details ird
        WHERE ird.sales_order_details_id = @order_details_id;

        IF @rel_required IS NOT NULL
        BEGIN
            IF ISNULL(@rel_released, 0) >= @rel_required
                SET @new_status = 'SCHEDULED DISPATCH - READY';
            ELSE
                SET @new_status = 'SCHEDULED DISPATCH - PREP';
        END
    END

    -- 3. Pick Activity (§7.1 rows 11-12) - same qty-based first/second-half pattern.
    IF @new_status IS NULL
    BEGIN
        DECLARE @pick_qty BIGINT, @actual_qty BIGINT;

        SELECT @pick_qty = SUM(pad.pick_qty), @actual_qty = SUM(ISNULL(pad.actual_qty, 0))
        FROM tbl_inv_pick_activity_details2 pad
        WHERE pad.sales_order_details_id = @order_details_id;

        IF @pick_qty IS NOT NULL
        BEGIN
            IF ISNULL(@actual_qty, 0) >= @pick_qty
                SET @new_status = 'PREPARED';
            ELSE
                SET @new_status = 'PREPARING';
        END
    END

    -- 4. Production / Job Order (§7.1 rows 5-10). is_accepted/is_wh_acknowledged are the
    -- fields SO_Item_Status_Module_Spec_2026-08-13.md identified as missing and added
    -- this pass (Phase 2 item 2.5) - without them rows 5/6 and 9/10 could not be told
    -- apart at all.
    IF @new_status IS NULL
    BEGIN
        DECLARE @jo_status NVARCHAR(100), @jo_is_accepted BIT, @jo_engr_id BIGINT,
                @jo_item_rqst NVARCHAR(MAX), @jo_wh_ack BIT;

        SELECT TOP 1
            @jo_status = jo.status,
            @jo_is_accepted = jo.is_accepted,
            @jo_engr_id = jo.engr_id,
            @jo_item_rqst = jo.item_rqst,
            @jo_wh_ack = jo.is_wh_acknowledged
        FROM tbl_trans_job_order jo
        WHERE jo.order_details_id = @order_details_id
        ORDER BY jo.id DESC;

        IF @jo_status IS NOT NULL OR @jo_is_accepted IS NOT NULL
        BEGIN
            IF UPPER(LTRIM(RTRIM(ISNULL(@jo_status, '')))) = 'COMPLETE' AND ISNULL(@jo_wh_ack, 0) = 1
                SET @new_status = 'IN STOCK';
            ELSE IF UPPER(LTRIM(RTRIM(ISNULL(@jo_status, '')))) = 'COMPLETE'
                SET @new_status = 'CHECKING';
            ELSE IF ISNULL(@jo_item_rqst, '') <> ''
                SET @new_status = 'IN PRODUCTION';
            ELSE IF ISNULL(@jo_engr_id, 0) > 0
                SET @new_status = 'PREPARING FOR PRODUCTION';
            ELSE IF ISNULL(@jo_is_accepted, 0) = 1
                SET @new_status = 'WAITING FOR ENGR';
            ELSE
                SET @new_status = 'WAITING ACKNOWLEDGEMENT';
        END
    END

    -- 5. Purchasing PO / RR (§7.1 rows 3-4). order_detail_ids is a comma-joined string
    -- of SO detail ids (one PO line can consolidate the same item across several SOs,
    -- per CLAUDE.md invariant #7) - matched with LIKE rather than STRING_SPLIT, which
    -- this DB's compatibility level (110) doesn't support.
    IF @new_status IS NULL
    BEGIN
        DECLARE @po_details_id BIGINT;

        SELECT TOP 1 @po_details_id = pod.id
        FROM tbl_purchasing_purchase_order_details pod
        WHERE ',' + ISNULL(pod.order_detail_ids, '') + ','
              LIKE '%,' + CAST(@order_details_id AS NVARCHAR(20)) + ',%'
        ORDER BY pod.id DESC;

        IF @po_details_id IS NOT NULL
        BEGIN
            DECLARE @received_qty BIGINT;

            SELECT @received_qty = SUM(ISNULL(rrd.received_qty, 0))
            FROM tbl_inv_receiving_report_details rrd
            WHERE rrd.purchase_order_details_id = @po_details_id;

            IF ISNULL(@received_qty, 0) >= ISNULL(@required_qty, 0)
                SET @new_status = 'IN STOCK';
            ELSE
                SET @new_status = 'WAITING FOR DELIVERY';
        END
    END

    -- 6. Base: stock availability + this SO's own reservations (§7.1 rows 1-2).
    -- Reservations MUST be added back for the SO that holds them (§7.1's own caveat) -
    -- keyed on quotation_id + item_id since tbl_inv_stock_reservations has no direct
    -- sales_order_details_id.
    IF @new_status IS NULL
    BEGIN
        DECLARE @stock_qty BIGINT, @reserved_qty BIGINT;

        SELECT @stock_qty = SUM(ISNULL(stk.stock_qty, 0))
        FROM tbl_inv_item_stocks stk
        WHERE stk.item_id = @item_id;

        SELECT @reserved_qty = SUM(r.qty)
        FROM tbl_inv_stock_reservations r
        WHERE r.item_id = @item_id
          AND r.quotation_id = @quotation_id
          AND r.status = 'Approved'
          AND (r.expires_at IS NULL OR r.expires_at > SYSDATETIMEOFFSET());

        IF (ISNULL(@stock_qty, 0) + ISNULL(@reserved_qty, 0)) >= ISNULL(@required_qty, 0)
            SET @new_status = 'IN STOCK';
        ELSE
            SET @new_status = 'CANVASS';
    END

    UPDATE tbl_trans_sales_order_details
    SET status = @new_status
    WHERE order_details_id = @order_details_id
      AND ISNULL(status, '') <> @new_status;
END;
