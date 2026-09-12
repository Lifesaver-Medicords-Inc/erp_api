package billing_services

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/initializers"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/models/accounting_models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// A/R Record of Transactions (spec 12.3) and the Billing list (12.5), plus the
// SO's hand-keyed billing ledger (12.4).
//
// Balances are derived, never stored: an invoice's balance is its total less
// what Payment Receipts applied to it (amount applied plus withholding
// applied), and an SO's balance is its invoices' total, plus its ledger, less
// those payments. NEXT DUE is the one typed field (12.3) and lives on the
// invoice.
type BillingService struct{}

func NewBillingService() *BillingService {
	return &BillingService{}
}

// The SO reference is typed on the invoice and may read "SO#0012", "SO 0012" or
// "0012", while the order itself stores its own document number. Both sides are
// reduced to digits before they are compared, in SQL and in Go, so the join does
// not depend on how it was keyed.
const soKeySQL = `REPLACE(REPLACE(REPLACE(UPPER(LTRIM(RTRIM(ISNULL(%s, '')))), 'SO#', ''), 'SO', ''), ' ', '')`

func soKey(value string) string {
	v := strings.ToUpper(strings.TrimSpace(value))
	v = strings.ReplaceAll(v, "SO#", "")
	v = strings.ReplaceAll(v, "SO", "")
	return strings.ReplaceAll(v, " ", "")
}

func key(column string) string {
	return strings.Replace(soKeySQL, "%s", column, 1)
}

// ---------------------------------------------------------------- 12.3 A/R

func (s *BillingService) GetARRecords() ([]accounting_models.ARRecordRow, int, error) {
	rows := []accounting_models.ARRecordRow{}

	q := `
SELECT si.id AS sales_invoice_id,
       '' AS tag,
       ISNULL(si.customer, '') AS customer,
       ISNULL(si.customer_code, '') AS customer_code,
       'SI#' + RIGHT('0000' + CAST(ISNULL(si.doc_no, 0) AS NVARCHAR(20)), 4) AS sales_invoice_no,
       ISNULL(si.doc_date, '') AS doc_date,
       ISNULL(si.reference_doc_so, '') AS so_no,
       ISNULL(si.payment_term, '') AS payment_term,
       ISNULL(si.total_amount_due, 0) AS total_amount_due,
       ISNULL(p.applied, 0) AS paid,
       ISNULL(si.total_amount_due, 0) - ISNULL(p.applied, 0) AS balance,
       ISNULL(si.next_due, '') AS next_due
FROM tbl_accounting_sales_invoice si
LEFT JOIN (
    SELECT d.sales_invoice_id,
           SUM(ISNULL(d.amount_applied, 0) + ISNULL(d.twas_applied, 0)) AS applied
    FROM tbl_accounting_payment_receipt_details d
    GROUP BY d.sales_invoice_id
) p ON p.sales_invoice_id = si.id
ORDER BY si.doc_no DESC`

	if err := initializers.DB.Raw(q).Scan(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed reading the A/R record of transactions")
	}

	for i := range rows {
		rows[i].Tag = tag(rows[i].Balance)
	}

	return rows, fiber.StatusOK, nil
}

// SetNextDue stores the date A/R typed against an invoice. Blank clears it, so
// the column is written with UpdateColumns rather than a struct update, which
// would skip an empty value.
func (s *BillingService) SetNextDue(salesInvoiceId uint, nextDue string, at models.At) (int, error) {
	var invoices []accounting_models.SalesInvoice
	if err := initializers.DB.Where("id = ?", salesInvoiceId).Limit(1).Find(&invoices).Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed reading the sales invoice")
	}
	if len(invoices) == 0 {
		return fiber.StatusNotFound, errors.New("sales invoice not found")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := tx.Model(&accounting_models.SalesInvoice{}).
		Where("id = ?", salesInvoiceId).
		UpdateColumns(map[string]interface{}{"next_due": nextDue}).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed setting next due")
	}

	invoice := invoices[0]
	invoice.NextDue = nextDue
	audit := accounting_models.SalesInvoiceAt{
		RefId:               invoice.ID,
		SalesInvoiceContent: invoice.SalesInvoiceContent,
		At:                  at,
	}
	if err := services.DbInsert(tx, &audit); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed writing the sales invoice audit row")
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

// ---------------------------------------------------------------- 12.5 Billing

func (s *BillingService) GetBilling() ([]accounting_models.BillingRow, int, error) {
	rows := []accounting_models.BillingRow{}

	q := `
WITH inv AS (
    SELECT ` + key("si.reference_doc_so") + ` AS so_key,
           SUM(ISNULL(si.total_amount_due, 0)) AS total_due,
           MIN(NULLIF(ISNULL(si.next_due, ''), '')) AS next_due
    FROM tbl_accounting_sales_invoice si
    GROUP BY ` + key("si.reference_doc_so") + `
),
pay AS (
    SELECT ` + key("si.reference_doc_so") + ` AS so_key,
           SUM(ISNULL(d.amount_applied, 0) + ISNULL(d.twas_applied, 0)) AS paid
    FROM tbl_accounting_payment_receipt_details d
    JOIN tbl_accounting_sales_invoice si ON si.id = d.sales_invoice_id
    GROUP BY ` + key("si.reference_doc_so") + `
),
led AS (
    SELECT ` + key("t.so_no") + ` AS so_key,
           SUM(ISNULL(t.amount, 0)) AS adjustments
    FROM tbl_accounting_so_billing_transaction t
    GROUP BY ` + key("t.so_no") + `
)
SELECT ISNULL(so.document_no, '') AS so_no,
       ISNULL(so.date, '') AS date,
       ISNULL(so.project_name, '') AS project_name,
       ISNULL(so.sales_executive, '') AS sales_executive,
       ISNULL(so.customer_name, '') AS customer,
       ISNULL(inv.total_due, 0) AS total_due,
       ISNULL(led.adjustments, 0) AS adjustments,
       ISNULL(pay.paid, 0) AS paid,
       ISNULL(inv.total_due, 0) + ISNULL(led.adjustments, 0) - ISNULL(pay.paid, 0) AS balance,
       ISNULL(inv.next_due, '') AS next_due,
       '' AS status
FROM tbl_trans_sales_order so
LEFT JOIN inv ON inv.so_key = ` + key("so.document_no") + `
LEFT JOIN pay ON pay.so_key = ` + key("so.document_no") + `
LEFT JOIN led ON led.so_key = ` + key("so.document_no") + `
WHERE inv.so_key IS NOT NULL OR led.so_key IS NOT NULL
ORDER BY so.date DESC, so.document_no DESC`

	if err := initializers.DB.Raw(q).Scan(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed reading the billing list")
	}

	receipts, status, err := s.receipts()
	if err != nil {
		return rows, status, err
	}

	for i := range rows {
		// A row closes only when the balance reaches zero - not on the first
		// Official Receipt, however many a partly-paid SO already carries (14.37).
		if rows[i].Balance <= 0.005 {
			rows[i].Status = "CLOSED"
		} else {
			rows[i].Status = "ACTIVE"
		}
		rows[i].Receipts = receipts[soKey(rows[i].SoNo)]
	}

	return rows, fiber.StatusOK, nil
}

// Every Official Receipt recorded against an SO's invoices, keyed by SO.
func (s *BillingService) receipts() (map[string][]accounting_models.BillingReceipt, int, error) {
	list := []accounting_models.BillingReceipt{}

	q := `
SELECT ` + key("si.reference_doc_so") + ` AS so_no,
       ISNULL(pr.reference_or_no, '') AS or_no,
       ISNULL(NULLIF(pr.date_collect, ''), ISNULL(pr.doc_date, '')) AS date,
       SUM(ISNULL(d.amount_applied, 0) + ISNULL(d.twas_applied, 0)) AS amount,
       ISNULL(pr.doc_no, 0) AS doc_no
FROM tbl_accounting_payment_receipt_details d
JOIN tbl_accounting_payment_receipt pr ON pr.id = d.payment_receipt_id
JOIN tbl_accounting_sales_invoice si ON si.id = d.sales_invoice_id
WHERE ISNULL(pr.reference_or_no, '') <> ''
GROUP BY ` + key("si.reference_doc_so") + `,
         ISNULL(pr.reference_or_no, ''),
         ISNULL(NULLIF(pr.date_collect, ''), ISNULL(pr.doc_date, '')),
         ISNULL(pr.doc_no, 0)
ORDER BY 3`

	if err := initializers.DB.Raw(q).Scan(&list).Error; err != nil {
		return nil, fiber.StatusInternalServerError, errors.New("failed reading the official receipts")
	}

	bySo := map[string][]accounting_models.BillingReceipt{}
	for _, r := range list {
		bySo[r.SoNo] = append(bySo[r.SoNo], r)
	}

	return bySo, fiber.StatusOK, nil
}

// ---------------------------------------------------------------- 12.4 ledger

func (s *BillingService) GetTransactions(soNo string) ([]accounting_models.SoBillingTransaction, int, error) {
	rows := []accounting_models.SoBillingTransaction{}

	query := initializers.DB.Model(&accounting_models.SoBillingTransaction{})
	if strings.TrimSpace(soNo) != "" {
		query = query.Where(key("so_no")+" = ?", soKey(soNo))
	}

	if err := query.Order("date").Find(&rows).Error; err != nil {
		return rows, fiber.StatusInternalServerError, errors.New("failed reading the billing ledger")
	}

	return rows, fiber.StatusOK, nil
}

func (s *BillingService) CreateTransaction(row *accounting_models.SoBillingTransaction, at models.At) (int, error) {
	if strings.TrimSpace(row.SoNo) == "" {
		return fiber.StatusBadRequest, errors.New("the ledger row needs an SO number")
	}
	if strings.TrimSpace(row.TransactionType) == "" {
		return fiber.StatusBadRequest, errors.New("the ledger row needs a transaction type")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbInsert(tx, row); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed creating the ledger row")
	}
	if err := s.audit(tx, row, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

func (s *BillingService) UpdateTransaction(row *accounting_models.SoBillingTransaction, at models.At) (int, error) {
	if row.ID == 0 {
		return fiber.StatusBadRequest, errors.New("the ledger row needs an id")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	// Select("*") so a cleared description or remark is saved rather than skipped.
	if err := tx.Model(&accounting_models.SoBillingTransaction{}).
		Where("id = ?", row.ID).Select("*").Updates(row).Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed updating the ledger row")
	}
	if err := s.audit(tx, row, at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

func (s *BillingService) DeleteTransaction(id uint, at models.At) (int, error) {
	var rows []accounting_models.SoBillingTransaction
	if err := initializers.DB.Where("id = ?", id).Limit(1).Find(&rows).Error; err != nil {
		return fiber.StatusInternalServerError, errors.New("failed reading the ledger row")
	}
	if len(rows) == 0 {
		return fiber.StatusNotFound, errors.New("ledger row not found")
	}

	tx := initializers.DB.Begin()
	if tx.Error != nil {
		return fiber.StatusInternalServerError, errors.New("failed to start DB transaction")
	}

	if err := services.DbDelete(tx, &accounting_models.SoBillingTransaction{}, map[string]interface{}{"id": id}); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed deleting the ledger row")
	}
	if err := s.audit(tx, &rows[0], at); err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fiber.StatusInternalServerError, errors.New("failed to commit transaction")
	}

	return fiber.StatusOK, nil
}

func (s *BillingService) audit(tx *gorm.DB, row *accounting_models.SoBillingTransaction, at models.At) error {
	audit := accounting_models.SoBillingTransactionAt{
		RefId:                       row.ID,
		SoBillingTransactionContent: row.SoBillingTransactionContent,
		At:                          at,
	}
	if err := services.DbInsert(tx, &audit); err != nil {
		return errors.New("failed writing the ledger audit row")
	}
	return nil
}

func tag(balance float64) string {
	if balance <= 0.005 {
		return "PAID"
	}
	return "ONGOING"
}
