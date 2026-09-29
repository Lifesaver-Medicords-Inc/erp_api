//go:build dbtest

package sales_services

import (
	"os"
	"testing"

	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
)

// writeChangedFields writes exactly the fields a save reports as changed: a new value, a
// deliberate zero/blank/false - and NOTHING else on the row, even where the row the client
// sent holds a different (stale) value. A field sent as null, and a key that is not a column,
// are left alone. Runs inside a transaction that is always rolled back. Run by hand:
//
//	go test -tags dbtest -run TestWriteChangedFields -v ./services/sales_services/
func TestWriteChangedFields(t *testing.T) {
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

	var content models.SalesProjectContent
	if err := tx.Last(&content).Error; err != nil {
		t.Skipf("no content row to test against: %v", err)
	}
	id := content.ContentID

	// The stored row: what the other app has just saved.
	if err := tx.Model(&models.SalesProjectContent{}).Where("content_id = ?", id).UpdateColumns(map[string]interface{}{
		"template_project_id": 7,
		"flow_id":             2,
		"is_wiring":           true,
		"application":         "OLD APPLICATION",
		"item_set_notes":      "clear me",
		"head":                "25 - SET BY THE OTHER APP",
		"flow":                "100",
	}).Error; err != nil {
		t.Fatal(err)
	}

	// The row this client sends: its own edits, plus a STALE head it never changed.
	sent := content
	sent.TemplateProjectId = 0
	unticked := false
	sent.IsWiring = &unticked
	sent.ItemSetNotes = ""
	sent.Application = "NEW APPLICATION"
	sent.Head = "10 - STALE COPY"
	sent.Flow = "999 - STALE COPY"

	changes := map[string]FieldChange[any]{
		"template_project_id": {OldValue: 7.0, NewValue: 0.0},                           // "-- No Template --"
		"is_wiring":           {OldValue: true, NewValue: false},                        // unticked
		"item_set_notes":      {OldValue: "clear me", NewValue: ""},                     // cleared text
		"application":         {OldValue: "OLD APPLICATION", NewValue: "NEW APPLICATION"}, // a new value
		"flow_id":             {OldValue: 2.0, NewValue: nil},                           // not sent: must survive
		"flow":                {OldValue: "100", NewValue: nil},                         // not sent: must survive
		"size_up":             {OldValue: "a", NewValue: ""},                            // not a column: ignored
	}
	if err := writeChangedFields(tx, &sent, "content_id", id, changes); err != nil {
		t.Fatalf("writeChangedFields: %v", err)
	}

	var got struct {
		TemplateProjectId uint
		FlowId            uint
		IsWiring          bool
		Application       string
		ItemSetNotes      string
		Head              string
		Flow              string
	}
	if err := tx.Raw(`SELECT ISNULL(template_project_id,0) AS template_project_id, ISNULL(flow_id,0) AS flow_id,
			ISNULL(is_wiring,0) AS is_wiring, ISNULL(application,'') AS application, ISNULL(item_set_notes,'') AS item_set_notes,
			ISNULL(head,'') AS head, ISNULL(flow,'') AS flow
		FROM tbl_trans_sales_project_content WHERE content_id = ?`, id).Scan(&got).Error; err != nil {
		t.Fatal(err)
	}

	check := func(what string, ok bool, detail string) {
		t.Helper()
		if ok {
			t.Logf("ok   %s", what)
		} else {
			t.Errorf("FAIL %s: %s", what, detail)
		}
	}
	check("explicit No Template written", got.TemplateProjectId == 0, "template_project_id is not 0")
	check("untick written", !got.IsWiring, "is_wiring is still true")
	check("cleared text written", got.ItemSetNotes == "", "item_set_notes = "+got.ItemSetNotes)
	check("new value written", got.Application == "NEW APPLICATION", "application = "+got.Application)
	check("field sent as null left alone", got.FlowId == 2, "flow_id changed")
	check("field sent as null left alone (text)", got.Flow == "100", "flow = "+got.Flow)
	check("the other app's HEAD survives a save that did not change it", got.Head == "25 - SET BY THE OTHER APP",
		"head = "+got.Head+" - the save wrote its stale copy")
}

// A content save syncs only the child list it changed. It used to replace BOTH Size Up and
// FINAL with whatever the saving client held - so Sales saving a FLOW edit wiped the Size Up an
// engineer had picked seconds before. Rolled back.
//
//	go test -tags dbtest -run TestContentSaveKeepsUntouchedSizeUp -v ./services/sales_services/
func TestContentSaveKeepsUntouchedSizeUp(t *testing.T) {
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

	var content models.SalesProjectContent
	if err := tx.Last(&content).Error; err != nil {
		t.Skipf("no content row to test against: %v", err)
	}

	// The engineer's pick, already saved.
	picked := models.SalesProjectSizeUp{}
	picked.SalesProjectContentID = content.ContentID
	picked.Model = "PICKED BY ENGINEERING"
	if err := tx.Create(&picked).Error; err != nil {
		t.Fatal(err)
	}
	count := func() int64 {
		var n int64
		tx.Model(&models.SalesProjectSizeUp{}).Where("sales_project_content_id = ?", content.ContentID).Count(&n)
		return n
	}
	before := count()

	// Sales saves FLOW, holding a stale (empty) Size Up list.
	sent := content
	sent.Flow = "123"
	sent.SalesProjectSizeUp = nil
	sent.SalesProjectContentFinal = nil
	diff := CollectionDiff[models.SalesProjectContent]{
		Updated: []UpdatedItem[models.SalesProjectContent]{{
			Item:    sent,
			Changes: map[string]FieldChange[any]{"flow": {OldValue: content.Flow, NewValue: "123"}},
		}},
	}
	if err := applyContentDiff(tx, content.BasedId, diff, models.At{}); err != nil {
		t.Fatalf("applyContentDiff: %v", err)
	}
	if after := count(); after != before {
		t.Errorf("a FLOW-only save changed the Size Up rows: %d before, %d after", before, after)
	} else {
		t.Logf("ok   FLOW-only save left the %d Size Up row(s) alone", after)
	}

	// A save that DID change Size Up still replaces it.
	diff.Updated[0].Changes = map[string]FieldChange[any]{"size_up": {OldValue: "x", NewValue: ""}}
	if err := applyContentDiff(tx, content.BasedId, diff, models.At{}); err != nil {
		t.Fatalf("applyContentDiff: %v", err)
	}
	if after := count(); after != 0 {
		t.Errorf("a save that emptied Size Up left %d row(s)", after)
	} else {
		t.Logf("ok   a save that changed Size Up replaced it")
	}
}
