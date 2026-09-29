//go:build dbtest

package sales_services

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// The quotation versioning rules (user decision 2026-09-25): Q#YYYY-nnnn.version.sub, New
// Version and Edit sessions numbered by the server, a replaced copy read-only, an FQ frozen,
// and FQ -> SO only from a finalized quote. Everything runs inside one transaction that is
// always rolled back. Run by hand:
//
//	go test -tags dbtest -run TestQuotationVersioning -v ./services/sales_services/
func TestQuotationVersioning(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	insert := func(q *models.SalesQuotation) {
		t.Helper()
		if err := tx.Create(q).Error; err != nil {
			t.Fatalf("insert %s: %v", q.DocumentNo, err)
		}
	}
	code := func(q models.SalesQuotation) string {
		return models.FormatQuoteCode(q.DocumentNo, q.VersionNo, q.SubVersionNo)
	}

	// ---- a new quote: next number in the year it was raised ----
	fresh := models.SalesQuotation{}
	fresh.Date = "2026-09-25 10:00:00"
	fresh.ProjectName = "versioning test"
	fresh.VersionNo, fresh.SubVersionNo = "1", "0"
	if err := assignNewQuotationDocNo(tx, &fresh); err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(fresh.DocumentNo, "Q#2026-") {
		t.Fatalf("new quote numbered %q, want Q#2026-nnnn", fresh.DocumentNo)
	}
	t.Logf("new quote:      %s", code(fresh))
	if code(fresh) != fresh.DocumentNo+".1.0" {
		t.Errorf("a new quote must be .1.0, got %s", code(fresh))
	}
	insert(&fresh)

	// ---- the next quote that year takes the following number ----
	second := models.SalesQuotation{}
	second.Date = "2026-10-01 08:00:00"
	second.VersionNo, second.SubVersionNo = "1", "0"
	if err := assignNewQuotationDocNo(tx, &second); err != nil {
		t.Fatal(err)
	}
	t.Logf("next quote:     %s", code(second))
	if second.DocumentNo == fresh.DocumentNo {
		t.Errorf("two new quotes got the same number %s", second.DocumentNo)
	}

	// ---- a quote raised in another year starts that year's count ----
	nextYear := models.SalesQuotation{}
	nextYear.Date = "2027-01-03 09:00:00"
	nextYear.VersionNo, nextYear.SubVersionNo = "1", "0"
	if err := assignNewQuotationDocNo(tx, &nextYear); err != nil {
		t.Fatal(err)
	}
	t.Logf("2027 quote:     %s", code(nextYear))
	if nextYear.DocumentNo != "Q#2027-0001" {
		t.Errorf("first quote of 2027 = %s, want Q#2027-0001 (the count restarts each January)", nextYear.DocumentNo)
	}

	// ---- New Version: same number, next version, sub 0 ----
	v2 := models.SalesQuotation{}
	v2.DocumentNo, v2.VersionNo, v2.SubVersionNo = fresh.DocumentNo, "1", "0"
	v2.ProjectName = "versioning test"
	v2.VersionAction = "new_version"
	if err := assignNewQuotationDocNo(tx, &v2); err != nil {
		t.Fatal(err)
	}
	t.Logf("new version:    %s", code(v2))
	if code(v2) != fresh.DocumentNo+".2.0" {
		t.Errorf("New Version gave %s, want %s.2.0", code(v2), fresh.DocumentNo)
	}
	insert(&v2)

	// ---- Edit sessions on version 2: .2.1, then .2.2 ----
	var subs []models.SalesQuotation
	for i := 0; i < 2; i++ {
		s := models.SalesQuotation{}
		s.DocumentNo, s.VersionNo, s.SubVersionNo = v2.DocumentNo, "2", "0"
		s.ProjectName = "versioning test"
		s.VersionAction = "new_sub_version"
		if err := assignNewQuotationDocNo(tx, &s); err != nil {
			t.Fatal(err)
		}
		insert(&s)
		subs = append(subs, s)
		t.Logf("edit session %d: %s", i+1, code(s))
	}
	if subs[0].SubVersionNo != "1" || subs[1].SubVersionNo != "2" {
		t.Errorf("edit sessions numbered %s, %s - want 1, 2", subs[0].SubVersionNo, subs[1].SubVersionNo)
	}

	// ---- a replaced copy is read-only; the newest one is not ----
	if status, err := guardProjectCopyIsCurrent(tx, subs[0].ID); err == nil {
		t.Errorf("a save to the replaced %s was allowed", code(subs[0]))
	} else {
		t.Logf("save to %s refused (%d): %v", code(subs[0]), status, err)
	}
	if _, err := guardProjectCopyIsCurrent(tx, subs[1].ID); err != nil {
		t.Errorf("a save to the current %s was refused: %v", code(subs[1]), err)
	}

	// ---- finalize: the FQ keeps number, version and sub-version, and is frozen ----
	fq := models.SalesQuotation{}
	fq.DocumentNo, fq.VersionNo, fq.SubVersionNo = "FQ#"+strings.TrimPrefix(subs[1].DocumentNo, "Q#"), "2", "2"
	fq.IsFinalized = true
	fq.ProjectName = "versioning test"
	if err := assignNewQuotationDocNo(tx, &fq); err != nil {
		t.Fatal(err)
	}
	insert(&fq)
	t.Logf("finalized:      %s", code(fq))
	if code(fq) != "FQ#"+strings.TrimPrefix(fresh.DocumentNo, "Q#")+".2.2" {
		t.Errorf("FQ is %s, want the sub-version it came from", code(fq))
	}
	if _, err := guardProjectCopyIsCurrent(tx, fq.ID); err == nil {
		t.Errorf("a save to the FQ %s was allowed", code(fq))
	}

	// ---- FQ -> SO ----
	draftOrder := models.Order{}
	draftOrder.Quotation_ID = subs[1].ID
	if _, err := prepareOrderFromFinalizedQuote(tx, &draftOrder); err == nil {
		t.Errorf("a sales order was allowed from the draft %s", code(subs[1]))
	} else {
		t.Logf("SO from a draft refused: %v", err)
	}

	order := models.Order{}
	order.Quotation_ID = fq.ID
	if _, err := prepareOrderFromFinalizedQuote(tx, &order); err != nil {
		t.Fatalf("SO from %s refused: %v", code(fq), err)
	}
	t.Logf("SO from FQ:     SO#%s, document_no %s", order.Doc, order.Document_No)
	if order.Document_No != code(fq) {
		t.Errorf("SO document_no = %s, want the FQ's full code %s", order.Document_No, code(fq))
	}
	if err := tx.Create(&order).Error; err != nil {
		t.Fatalf("insert SO: %v", err)
	}

	again := models.Order{}
	again.Quotation_ID = fq.ID
	if _, err := prepareOrderFromFinalizedQuote(tx, &again); err == nil {
		t.Errorf("a second live SO from %s was allowed", code(fq))
	} else {
		t.Logf("second SO refused: %v", err)
	}
}

// EDIT on a project quotation copies it onto the next sub-version (StartProjectSubVersion).
// Run against the busiest current project draft in the database, inside a transaction that is
// always rolled back:
//
//   - every table under the project has as many rows under the copy as under the source, and
//     the same values in every column but the keys (checksummed column by column);
//   - the header is the same apart from its number;
//   - reservations moved to the copy's rows; nothing left on the old ones;
//   - Change History copied, plus the "PROJECT - SUB-VERSION" line;
//   - the replaced copy refuses a second Edit and a save; the copy accepts both.
//
//	go test -tags dbtest -run TestProjectSubVersionCopy -v ./services/sales_services/
func TestProjectSubVersionCopy(t *testing.T) {
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()

	tx := initializers.DB.Begin()
	defer tx.Rollback()

	must := func(err error) {
		t.Helper()
		if err != nil {
			t.Fatal(err)
		}
	}

	var sourceID uint
	must(tx.Raw(`
		SELECT TOP 1 q.id
		  FROM tbl_trans_sales_quotation q
		 WHERE ISNULL(q.project_name, '') <> '' AND ISNULL(q.is_finalized, 0) = 0
		   AND NOT EXISTS (SELECT 1 FROM tbl_trans_sales_quotation n
		                    WHERE n.document_no = q.document_no AND n.version_no = q.version_no
		                      AND TRY_CONVERT(int, n.sub_version_no) > TRY_CONVERT(int, q.sub_version_no))
		 ORDER BY (SELECT COUNT(*) FROM tbl_trans_sales_project_items i
		             JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = i.based_id
		            WHERE s.based_id = q.id) DESC, q.id DESC`).Scan(&sourceID).Error)
	if sourceID == 0 {
		t.Skip("no current project draft in this database")
	}

	// Rows under one quotation, per table. {Q} is the quotation id.
	sets := `(SELECT item_set_id FROM tbl_trans_sales_project_item_set WHERE based_id = {Q})`
	contents := `(SELECT content_id FROM tbl_trans_sales_project_content WHERE based_id IN ` + sets + `)`
	items := `(SELECT items_id FROM tbl_trans_sales_project_items WHERE based_id IN ` + sets + `)`
	wirings := `(SELECT id FROM tbl_trans_sales_project_wiring WHERE based_id IN ` + sets + `)`
	under := map[string]string{
		"tbl_trans_sales_project_multiplier":                     `based_id = {Q}`,
		"tbl_trans_sales_project_item_set":                       `based_id = {Q}`,
		"tbl_trans_sales_project_content":                        `based_id IN ` + sets,
		"tbl_trans_sales_project_content_final":                  `sales_project_content_id IN ` + contents,
		"tbl_trans_sales_project_size_up":                        `sales_project_content_id IN ` + contents,
		"tbl_trans_sales_project_content_advanced_configuration": `based_id IN ` + sets,
		"tbl_trans_sales_project_items":                          `based_id IN ` + sets,
		"tbl_trans_sales_quotation_selected_image":               `quotation_quick_id IN ` + items,
		"tbl_trans_sales_project_wiring":                         `based_id IN ` + sets,
		"tbl_wiring_user_inputs":                                 `items_id IN ` + items + ` OR wiring_id IN ` + wirings,
	}
	q := func(sql string, id uint) string { return strings.ReplaceAll(sql, "{Q}", fmt.Sprint(id)) }

	// A quotation's history, told apart from same-numbered item sets' history by label - the
	// rule the apps use (ProjectHistoryRows) and the copy follows.
	historyOf := func(quotationID uint, itemSets string) string {
		return fmt.Sprintf(`SELECT COUNT(*) FROM tbl_trans_sales_project_history h
		  WHERE (h.based_id = %d AND NOT (h.old_data LIKE 'ITEM/SET %%' OR h.old_data LIKE 'Sub Project: %%'
		          OR h.old_data LIKE 'Advance Conditions: %%' OR h.old_data LIKE 'Project Items Table -> %%'
		          OR h.old_data LIKE 'Project Wiring Table -> %%'))
		     OR (h.based_id IN %s AND NOT (h.old_data LIKE 'PROJECT - %%' OR h.old_data LIKE 'Header -> %%'
		          OR h.old_data LIKE 'Footer -> %%' OR h.old_data LIKE 'Project Header: %%'
		          OR h.old_data LIKE 'Multiplier Table -> %%'))`, quotationID, itemSets)
	}

	// Count and a checksum of every column except the row's own id and the ids it points at.
	fingerprint := func(step projectCloneStep, quotationID uint) (int64, int64) {
		cols, err := writableColumns(tx, step.Table)
		must(err)
		var kept []string
		for _, c := range cols {
			if _, remapped := step.Remap[c]; !remapped {
				kept = append(kept, quoteIdent(c))
			}
		}
		sum := "0"
		if len(kept) > 0 {
			sum = "ISNULL(CHECKSUM_AGG(CHECKSUM(" + strings.Join(kept, ", ") + ")), 0)"
		}
		var r struct {
			N   int64
			Sum int64
		}
		must(tx.Raw(fmt.Sprintf(`SELECT COUNT(*) AS n, %s AS sum FROM %s WHERE %s`,
			sum, step.Table, q(under[step.Table], quotationID))).Scan(&r).Error)
		return r.N, r.Sum
	}

	// Two reserved lines on the source, so there is something to move.
	must(tx.Exec(`
		INSERT INTO tbl_inv_stock_reservations (item_id, qty, source_type, source_id, quotation_id, status)
		SELECT TOP 2 i.item_id, 1, ?, i.items_id, ?, 'Approved'
		  FROM tbl_trans_sales_project_items i
		  JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = i.based_id
		 WHERE s.based_id = ?
		 ORDER BY i.items_id`, projectItemReservationSource, sourceID, sourceID).Error)

	var reservedBefore, historyBefore int64
	must(tx.Raw(`SELECT COUNT(*) FROM tbl_inv_stock_reservations WHERE source_type = ? AND quotation_id = ?`,
		projectItemReservationSource, sourceID).Scan(&reservedBefore).Error)
	must(tx.Raw(historyOf(sourceID, q(sets, sourceID))).Scan(&historyBefore).Error)

	var source models.SalesQuotation
	must(tx.First(&source, sourceID).Error)

	note := &models.SalesProjectHistoryContent{User: "tester", Date: "9/25/2026", Time: "10:00 AM"}
	copyQ, status, err := StartProjectSubVersion(tx, sourceID, note, models.At{})
	if err != nil {
		t.Fatalf("Edit on %s refused (%d): %v", source.QuoteCode, status, err)
	}
	t.Logf("Edit on %s (id %d) -> %s (id %d)", source.QuoteCode, sourceID, copyQ.QuoteCode, copyQ.ID)

	wantSub := fmt.Sprint(atoi(source.SubVersionNo) + 1)
	if copyQ.VersionNo != source.VersionNo || copyQ.SubVersionNo != wantSub || copyQ.IsFinalized {
		t.Errorf("copy is %s, want version %s sub-version %s, draft", copyQ.QuoteCode, source.VersionNo, wantSub)
	}

	// Header: the same in every column but the number.
	cols, err := writableColumns(tx, "tbl_trans_sales_quotation")
	must(err)
	var cmp []string
	for _, c := range cols {
		switch c {
		case "document_no", "version_no", "sub_version_no", "is_finalized":
			continue
		}
		cmp = append(cmp, fmt.Sprintf("(a.%[1]s = b.%[1]s OR (a.%[1]s IS NULL AND b.%[1]s IS NULL))", quoteIdent(c)))
	}
	var headerSame int64
	must(tx.Raw(`SELECT COUNT(*) FROM tbl_trans_sales_quotation a JOIN tbl_trans_sales_quotation b ON b.id = ?
	              WHERE a.id = ? AND `+strings.Join(cmp, " AND "), copyQ.ID, sourceID).Scan(&headerSame).Error)
	if headerSame != 1 {
		t.Errorf("the copy's header differs from the source in a column other than its number")
	}

	for _, step := range projectCloneSteps {
		n1, s1 := fingerprint(step, sourceID)
		n2, s2 := fingerprint(step, copyQ.ID)
		t.Logf("%-56s source %4d  copy %4d  same values: %v", step.Table, n1, n2, s1 == s2)
		if n1 != n2 || s1 != s2 {
			t.Errorf("%s: source %d rows (checksum %d), copy %d rows (checksum %d)", step.Table, n1, s1, n2, s2)
		}
	}

	var onSource, onCopy, historyAfter int64
	must(tx.Raw(`SELECT COUNT(*) FROM tbl_inv_stock_reservations WHERE source_type = ? AND quotation_id = ?`,
		projectItemReservationSource, sourceID).Scan(&onSource).Error)
	must(tx.Raw(`SELECT COUNT(*) FROM tbl_inv_stock_reservations r
	               JOIN tbl_trans_sales_project_items i ON i.items_id = r.source_id
	               JOIN tbl_trans_sales_project_item_set s ON s.item_set_id = i.based_id AND s.based_id = ?
	              WHERE r.source_type = ? AND r.quotation_id = ?`,
		copyQ.ID, projectItemReservationSource, copyQ.ID).Scan(&onCopy).Error)
	t.Logf("reservations: %d before; after, %d on the old copy and %d on the new copy's rows", reservedBefore, onSource, onCopy)
	if onSource != 0 || onCopy != reservedBefore {
		t.Errorf("reservations did not move with the copy")
	}

	must(tx.Raw(historyOf(copyQ.ID, q(sets, copyQ.ID))).Scan(&historyAfter).Error)
	var boundary models.SalesProjectHistory
	must(tx.Where("based_id = ? AND old_data = ?", copyQ.ID, "PROJECT - SUB-VERSION").
		Order("history_id DESC").Limit(1).Find(&boundary).Error)
	t.Logf("history: source %d rows, copy %d rows; boundary %q", historyBefore, historyAfter, boundary.NewData)
	if historyAfter != historyBefore+1 || boundary.NewData != source.QuoteCode+" => "+copyQ.QuoteCode {
		t.Errorf("history not carried over with its boundary line")
	}

	if _, _, err := StartProjectSubVersion(tx, sourceID, nil, models.At{}); err == nil {
		t.Errorf("a second Edit on the replaced %s was allowed", source.QuoteCode)
	} else {
		t.Logf("second Edit on the replaced copy refused: %v", err)
	}
	if _, err := guardProjectCopyIsCurrent(tx, sourceID); err == nil {
		t.Errorf("a save to the replaced %s was allowed", source.QuoteCode)
	}
	if _, err := guardProjectCopyIsCurrent(tx, copyQ.ID); err != nil {
		t.Errorf("a save to the copy %s was refused: %v", copyQ.QuoteCode, err)
	}
	again, _, err := StartProjectSubVersion(tx, copyQ.ID, nil, models.At{})
	must(err)
	t.Logf("the next Edit gives %s", again.QuoteCode)
}

func atoi(s string) int {
	n, _ := strconv.Atoi(strings.TrimSpace(s))
	return n
}
