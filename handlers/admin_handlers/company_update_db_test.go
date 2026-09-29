//go:build dbtest

package adminhandlers

import (
	"bytes"
	"encoding/json"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	adminservices "github.com/pierceperado/smpc/services/admin_services"
)

// PUT /companies/:id adds an address and new contacts, updates existing ones in place, leaves
// both alone when the caller sends neither (Accounting's Company Setup), and never touches the
// company fields a caller did not send (VAT rate, fees).
//
// It commits (the company services open their own transactions), so it refuses to run on
// anything but a scratch database, and puts company 1 back as it found it:
//
//	DB_NAME=Lightspeed_rehearsal_controllers_test go test -tags dbtest ./handlers/admin_handlers/ -run TestCompanyUpdate -v
func TestCompanyUpdateAddressAndContacts(t *testing.T) {
	if !strings.HasSuffix(os.Getenv("DB_NAME"), "_test") {
		t.Skip("set DB_NAME to a scratch *_test database - this test commits")
	}
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()
	db := initializers.DB

	var before struct {
		Inclusions string  `gorm:"column:inclusions_quotation_terms"`
		Terms      string  `gorm:"column:term_and_conditions"`
		Vat        float64 `gorm:"column:vat_rate_percent"`
	}
	db.Raw(`SELECT inclusions_quotation_terms, term_and_conditions, ISNULL(vat_rate_percent,0) vat_rate_percent FROM tbl_company WHERE id = 1`).Scan(&before)
	var addrBefore, contactsBefore int64
	db.Raw(`SELECT COUNT(*) FROM tbl_company_address WHERE company_id = 1`).Scan(&addrBefore)
	db.Raw(`SELECT COUNT(*) FROM tbl_company_contact WHERE company_id = 1`).Scan(&contactsBefore)
	if addrBefore != 0 || contactsBefore != 0 {
		t.Fatalf("expected company 1 with no address/contacts on the scratch DB, found %d/%d", addrBefore, contactsBefore)
	}
	t.Cleanup(func() {
		db.Exec(`DELETE FROM z_tbl_company_contact_at WHERE ref_id IN (SELECT id FROM tbl_company_contact WHERE company_id = 1)`)
		db.Exec(`DELETE FROM tbl_company_contact WHERE company_id = 1`)
		db.Exec(`DELETE FROM z_tbl_company_address_at WHERE ref_id IN (SELECT id FROM tbl_company_address WHERE company_id = 1)`)
		db.Exec(`DELETE FROM tbl_company_address WHERE company_id = 1`)
		db.Exec(`UPDATE tbl_company SET inclusions_quotation_terms = ? WHERE id = 1`, before.Inclusions)
	})

	h := NewCompanyHandler(adminservices.NewCompanyService())
	app := fiber.New()
	app.Put("/companies/:id", func(c *fiber.Ctx) error {
		c.Locals("at", models.At{AtAction: "TEST", AtUser: "dbtest"})
		return h.UpdateCompanyHandler(c)
	})
	put := func(body map[string]interface{}) {
		t.Helper()
		raw, _ := json.Marshal(body)
		req := httptest.NewRequest("PUT", "/companies/1", bytes.NewReader(raw))
		req.Header.Set("Content-Type", "application/json")
		res, err := app.Test(req, 30000)
		if err != nil {
			t.Fatal(err)
		}
		var out map[string]interface{}
		json.NewDecoder(res.Body).Decode(&out)
		if res.StatusCode != 200 {
			t.Fatalf("PUT returned %d: %v", res.StatusCode, out)
		}
	}
	count := func(q string) int64 {
		var n int64
		db.Raw(q).Scan(&n)
		return n
	}
	text := func(q string) string {
		var s string
		db.Raw(q).Scan(&s)
		return s
	}

	// 1. What the Admin page sends for company 1: a new address and a new contact.
	put(map[string]interface{}{
		"id": 1, "company_code": "0001", "inclusions_quotation_terms": "TEST INCLUSIONS",
		"address":  map[string]interface{}{"address_type": "Main", "city": "Makati", "postal_code": 1200},
		"contacts": []map[string]interface{}{{"id": 0, "full_name": "Juan Cruz", "designation": "GM", "is_primary_contact": true}},
	})
	if n := count(`SELECT COUNT(*) FROM tbl_company_address WHERE company_id = 1`); n != 1 {
		t.Fatalf("address rows after first save = %d, want 1 (created)", n)
	}
	if c := text(`SELECT city FROM tbl_company_address WHERE company_id = 1`); c != "Makati" {
		t.Errorf("city = %q, want Makati", c)
	}
	if n := count(`SELECT COUNT(*) FROM tbl_company_contact WHERE company_id = 1`); n != 1 {
		t.Fatalf("contacts after first save = %d, want 1 (created)", n)
	}
	if s := text(`SELECT inclusions_quotation_terms FROM tbl_company WHERE id = 1`); s != "TEST INCLUSIONS" {
		t.Errorf("inclusions = %q, want TEST INCLUSIONS", s)
	}

	// 2. Edit again: the address is updated in place, the saved contact updated, a second added.
	var contactID int64
	db.Raw(`SELECT TOP 1 id FROM tbl_company_contact WHERE company_id = 1`).Scan(&contactID)
	put(map[string]interface{}{
		"id": 1, "company_code": "0001",
		"address": map[string]interface{}{"city": "Taguig"},
		"contacts": []map[string]interface{}{
			{"id": contactID, "full_name": "Juan Cruz", "designation": "President", "is_primary_contact": true},
			{"id": 0, "full_name": "Maria Santos", "designation": "Accountant", "is_primary_contact": false},
		},
	})
	if n := count(`SELECT COUNT(*) FROM tbl_company_address WHERE company_id = 1`); n != 1 {
		t.Errorf("address rows after second save = %d, want still 1 (updated, not duplicated)", n)
	}
	if c := text(`SELECT city FROM tbl_company_address WHERE company_id = 1`); c != "Taguig" {
		t.Errorf("city = %q, want Taguig", c)
	}
	if n := count(`SELECT COUNT(*) FROM tbl_company_contact WHERE company_id = 1`); n != 2 {
		t.Errorf("contacts after second save = %d, want 2", n)
	}
	if d := text(`SELECT designation FROM tbl_company_contact WHERE id = ` + itoa(contactID)); d != "President" {
		t.Errorf("updated contact designation = %q, want President", d)
	}

	// 3. What Accounting's Company Setup sends: no address, no contacts - neither changes.
	put(map[string]interface{}{"id": 1, "company_code": "0001", "company_name": text(`SELECT company_name FROM tbl_company WHERE id = 1`)})
	if n := count(`SELECT COUNT(*) FROM tbl_company_address WHERE company_id = 1`); n != 1 {
		t.Errorf("address rows after an address-less save = %d, want 1", n)
	}
	if n := count(`SELECT COUNT(*) FROM tbl_company_contact WHERE company_id = 1`); n != 2 {
		t.Errorf("contacts after a contact-less save = %d, want 2", n)
	}

	// Fields no caller above sent are untouched.
	var vat float64
	db.Raw(`SELECT ISNULL(vat_rate_percent,0) FROM tbl_company WHERE id = 1`).Scan(&vat)
	if vat != before.Vat {
		t.Errorf("vat_rate_percent = %v, want unchanged %v", vat, before.Vat)
	}
	if s := text(`SELECT term_and_conditions FROM tbl_company WHERE id = 1`); s != before.Terms {
		t.Errorf("term_and_conditions changed although no save sent it")
	}
}

func itoa(n int64) string {
	b, _ := json.Marshal(n)
	return string(b)
}

// POST /companies stores the address once and each contact once. GORM used to insert them with
// the company and the handler inserted them again.
func TestCompanyCreateStoresAddressAndContactsOnce(t *testing.T) {
	if !strings.HasSuffix(os.Getenv("DB_NAME"), "_test") {
		t.Skip("set DB_NAME to a scratch *_test database - this test commits")
	}
	if _, err := os.Stat(".env"); err != nil {
		if err := os.Chdir("../.."); err != nil {
			t.Fatal(err)
		}
	}
	initializers.LoadEnv()
	initializers.ConnectDb()
	initializers.InitRedis()
	db := initializers.DB

	h := NewCompanyHandler(adminservices.NewCompanyService())
	app := fiber.New()
	app.Post("/companies", func(c *fiber.Ctx) error {
		c.Locals("at", models.At{AtAction: "TEST", AtUser: "dbtest"})
		return h.CreateCompanyHandler(c)
	})
	raw, _ := json.Marshal(map[string]interface{}{
		"company_code": "ZZTEST", "company_name": "DBTEST COMPANY", "status": "Active",
		"term_and_conditions": "TEST T&C",
		"address": map[string]interface{}{"address_type": "Main", "city": "Pasig"},
		"contacts": []map[string]interface{}{
			{"full_name": "A One", "is_primary_contact": true},
			{"full_name": "B Two", "is_primary_contact": false},
		},
	})
	req := httptest.NewRequest("POST", "/companies", bytes.NewReader(raw))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req, 30000)
	if err != nil {
		t.Fatal(err)
	}
	var id int64
	db.Raw(`SELECT TOP 1 id FROM tbl_company WHERE company_code = 'ZZTEST' ORDER BY id DESC`).Scan(&id)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM z_tbl_company_contact_at WHERE ref_id IN (SELECT id FROM tbl_company_contact WHERE company_id = ?)`, id)
		db.Exec(`DELETE FROM tbl_company_contact WHERE company_id = ?`, id)
		db.Exec(`DELETE FROM z_tbl_company_address_at WHERE ref_id IN (SELECT id FROM tbl_company_address WHERE company_id = ?)`, id)
		db.Exec(`DELETE FROM tbl_company_address WHERE company_id = ?`, id)
		db.Exec(`DELETE FROM z_tbl_company_at WHERE ref_id = ?`, id)
		db.Exec(`DELETE FROM tbl_company WHERE id = ?`, id)
	})
	if res.StatusCode != 200 && res.StatusCode != 201 {
		t.Fatalf("POST returned %d", res.StatusCode)
	}
	if id == 0 {
		t.Fatal("company not created")
	}
	var addresses, contacts int64
	db.Raw(`SELECT COUNT(*) FROM tbl_company_address WHERE company_id = ?`, id).Scan(&addresses)
	db.Raw(`SELECT COUNT(*) FROM tbl_company_contact WHERE company_id = ?`, id).Scan(&contacts)
	if addresses != 1 {
		t.Errorf("address rows = %d, want 1", addresses)
	}
	if contacts != 2 {
		t.Errorf("contact rows = %d, want 2", contacts)
	}
	var terms string
	db.Raw(`SELECT term_and_conditions FROM tbl_company WHERE id = ?`, id).Scan(&terms)
	if terms != "TEST T&C" {
		t.Errorf("term_and_conditions = %q, want TEST T&C", terms)
	}
}
