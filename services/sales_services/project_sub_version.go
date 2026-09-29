package sales_services

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pierceperado/smpc/models"
	"github.com/pierceperado/smpc/services"
	"gorm.io/gorm"
)

// Edit on a project quotation starts the next sub-version: Q#2026-0001.2.0 -> Q#2026-0001.2.1
// (user decision 2026-09-25, "Clicking EDIT starts the next sub-version (a copy). Autosave keeps
// updating that one sub-version until you leave edit mode.").
//
// The copy is made HERE, from what is stored, before anything has been typed - not by the app
// posting back what its screen holds. Two reasons:
//
//   - Faithful. Every column of every row is copied, whether or not a Go or C# model knows
//     about it. The screen round-trip has lost data before (ids the C# models did not carry,
//     zeros the update path skipped), and a copy is the one place that must lose nothing.
//   - Nothing to race. The user has not typed yet, so there is nothing on screen that the
//     reload onto the new copy could throw away; every later autosave diffs against the new
//     copy exactly as it did against the old.
//
// Copied: the header, multipliers, tabs (item sets), each tab's content with its Size Up and
// Final Selection, advanced conditions, items with their selected images, wiring with its
// notes, and the Change History (plus a line marking the new sub-version). MOVED, not copied:
// stock reservations - the units stay held once, by the copy now being worked on.

// projectCloneStep is one table in the copy. Its rows are the ones whose Remap columns point
// at something already copied; those columns are rewritten to the copies' ids. Kind, when set,
// records old -> new ids of this table's own rows for the steps after it.
type projectCloneStep struct {
	Table string
	IDCol string
	Kind  string
	Remap map[string]string // column -> kind of the id it holds
}

var projectCloneSteps = []projectCloneStep{
	{Table: "tbl_trans_sales_project_multiplier", IDCol: "multiplier_id", Remap: map[string]string{"based_id": "quotation"}},
	{Table: "tbl_trans_sales_project_item_set", IDCol: "item_set_id", Kind: "item_set", Remap: map[string]string{"based_id": "quotation"}},
	{Table: "tbl_trans_sales_project_content", IDCol: "content_id", Kind: "content", Remap: map[string]string{"based_id": "item_set"}},
	{Table: "tbl_trans_sales_project_content_final", IDCol: "id", Remap: map[string]string{"sales_project_content_id": "content"}},
	{Table: "tbl_trans_sales_project_size_up", IDCol: "id", Remap: map[string]string{"sales_project_content_id": "content"}},
	{Table: "tbl_trans_sales_project_content_advanced_configuration", IDCol: "conditions_id", Remap: map[string]string{"based_id": "item_set"}},
	{Table: "tbl_trans_sales_project_items", IDCol: "items_id", Kind: "items", Remap: map[string]string{"based_id": "item_set"}},
	{Table: "tbl_trans_sales_quotation_selected_image", IDCol: "id", Remap: map[string]string{"quotation_quick_id": "items"}},
	{Table: "tbl_trans_sales_project_wiring", IDCol: "id", Kind: "wiring", Remap: map[string]string{"based_id": "item_set"}},
	{Table: "tbl_wiring_user_inputs", IDCol: "note_id", Remap: map[string]string{"items_id": "items", "wiring_id": "wiring"}},
}

// StartProjectSubVersion copies project quotation sourceID onto its version's next sub-version
// and returns the new copy. note, when sent, is who started the Edit session and when by their
// own clock; it becomes the "PROJECT - SUB-VERSION" line in Change History.
func StartProjectSubVersion(tx *gorm.DB, sourceID uint, note *models.SalesProjectHistoryContent, at models.At) (models.SalesQuotation, int, error) {
	var created models.SalesQuotation

	// Only the newest copy of a version is edited; an FQ never is (spec §5.2).
	if status, err := guardProjectCopyIsCurrent(tx, sourceID); err != nil {
		return created, status, err
	}

	var source models.SalesQuotation
	if err := tx.First(&source, sourceID).Error; err != nil {
		return created, fiber.StatusNotFound, errors.New("quotation not found")
	}
	if strings.TrimSpace(source.ProjectName) == "" {
		return created, fiber.StatusBadRequest, errors.New("not a project quotation")
	}

	next := source
	if err := assignNextSubVersion(tx, &next); err != nil {
		return created, fiber.StatusInternalServerError, err
	}

	// Created without parameters so it lives at session level: a statement with parameters
	// runs through sp_executesql, and a temp table created in there is gone when it returns.
	// The transaction keeps every later statement on this same connection.
	if err := tx.Exec(`
		IF OBJECT_ID('tempdb..#project_clone_map') IS NOT NULL DROP TABLE #project_clone_map;
		CREATE TABLE #project_clone_map (kind varchar(20) NOT NULL, old_id bigint NOT NULL, new_id bigint NOT NULL,
		                                 PRIMARY KEY (kind, old_id));`).Error; err != nil {
		return created, fiber.StatusInternalServerError, fmt.Errorf("preparing the copy: %w", err)
	}
	defer tx.Exec(`IF OBJECT_ID('tempdb..#project_clone_map') IS NOT NULL DROP TABLE #project_clone_map;`)

	newID, err := cloneQuotationHeader(tx, sourceID, next)
	if err != nil {
		return created, fiber.StatusInternalServerError, err
	}
	if err := tx.Exec(`INSERT INTO #project_clone_map (kind, old_id, new_id) VALUES ('quotation', ?, ?)`, sourceID, newID).Error; err != nil {
		return created, fiber.StatusInternalServerError, fmt.Errorf("preparing the copy: %w", err)
	}

	for _, step := range projectCloneSteps {
		if err := cloneProjectRows(tx, step); err != nil {
			return created, fiber.StatusInternalServerError, err
		}
	}

	// The units stay reserved, now by the copy being worked on - nothing is released and
	// nothing reserved again, so stock never moves.
	if err := tx.Exec(`
		UPDATE r SET r.source_id = m.new_id, r.quotation_id = ?
		  FROM tbl_inv_stock_reservations r
		  JOIN #project_clone_map m ON m.kind = 'items' AND m.old_id = r.source_id
		 WHERE r.source_type = ?`, newID, projectItemReservationSource).Error; err != nil {
		return created, fiber.StatusInternalServerError, fmt.Errorf("moving reservations to the new copy: %w", err)
	}

	if err := copyClonedProjectHistory(tx, sourceID, newID); err != nil {
		return created, fiber.StatusInternalServerError, err
	}

	if err := tx.First(&created, newID).Error; err != nil {
		return created, fiber.StatusInternalServerError, errors.New("failed reading the new copy")
	}

	if note != nil {
		boundary := *note
		boundary.OldData = "PROJECT - SUB-VERSION"
		boundary.NewData = models.FormatQuoteCode(source.DocumentNo, source.VersionNo, source.SubVersionNo) +
			" => " + created.QuoteCode
		if err := CreateSalesProjectHistory(tx, newID, models.SalesProjectHistory{SalesProjectHistoryContent: boundary}, at); err != nil {
			return created, fiber.StatusInternalServerError, err
		}
	}

	if err := services.DbInsert(tx, &models.SalesQuotationAt{RefId: newID, SalesQuotationContent: created.SalesQuotationContent, At: at}); err != nil {
		return created, fiber.StatusInternalServerError, errors.New("failed creating sales quotation at")
	}

	return created, 0, nil
}

// cloneQuotationHeader copies the header row under the new sub-version's number.
func cloneQuotationHeader(tx *gorm.DB, sourceID uint, next models.SalesQuotation) (uint, error) {
	cols, err := writableColumns(tx, "tbl_trans_sales_quotation")
	if err != nil {
		return 0, err
	}

	// Everything as the source has it, except the number and that it is a draft.
	overrides := map[string]interface{}{
		"document_no":    next.DocumentNo,
		"version_no":     next.VersionNo,
		"sub_version_no": next.SubVersionNo,
		"is_finalized":   false,
	}
	values := make([]string, len(cols))
	var args []interface{}
	for i, col := range cols {
		if v, ok := overrides[col]; ok {
			values[i] = "?"
			args = append(args, v)
		} else {
			values[i] = "s." + quoteIdent(col)
		}
	}
	args = append(args, sourceID)

	var newID uint
	if err := tx.Raw(fmt.Sprintf(`
		INSERT INTO tbl_trans_sales_quotation (%s)
		OUTPUT inserted.id
		SELECT %s FROM tbl_trans_sales_quotation s WHERE s.id = ?`,
		joinIdents(cols), strings.Join(values, ", ")), args...).Scan(&newID).Error; err != nil {
		return 0, fmt.Errorf("copying the quotation header: %w", err)
	}
	if newID == 0 {
		return 0, errors.New("copying the quotation header: no row written")
	}
	return newID, nil
}

// cloneProjectRows copies one table's rows that hang off rows already copied, pointing each
// copy at its parents' copies. MERGE rather than INSERT because only MERGE's OUTPUT can name
// the source row, which is how each old id is paired with its new one.
func cloneProjectRows(tx *gorm.DB, step projectCloneStep) error {
	cols, err := writableColumns(tx, step.Table)
	if err != nil {
		return err
	}

	var joins, matched []string
	values := make([]string, len(cols))
	alias := map[string]string{}
	n := 0
	for col, kind := range step.Remap {
		a := fmt.Sprintf("m%d", n)
		n++
		alias[col] = a
		joins = append(joins, fmt.Sprintf("LEFT JOIN #project_clone_map %s ON %s.kind = '%s' AND %s.old_id = s.%s",
			a, a, kind, a, quoteIdent(col)))
		matched = append(matched, a+".new_id IS NOT NULL")
	}

	selectCols := []string{"s.*"}
	for col, a := range alias {
		selectCols = append(selectCols, fmt.Sprintf("%s.new_id AS [__new_%s]", a, col))
	}
	for i, col := range cols {
		if _, ok := alias[col]; ok {
			// A reference to something that was not copied (a wiring note's item when only its
			// wiring was) keeps pointing where it did.
			values[i] = fmt.Sprintf("ISNULL(s.[__new_%s], s.%s)", col, quoteIdent(col))
		} else {
			values[i] = "s." + quoteIdent(col)
		}
	}

	output := ""
	if step.Kind != "" {
		output = fmt.Sprintf("OUTPUT '%s', s.%s, inserted.%s INTO #project_clone_map (kind, old_id, new_id)",
			step.Kind, quoteIdent(step.IDCol), quoteIdent(step.IDCol))
	}

	sql := fmt.Sprintf(`
		MERGE INTO %s AS t
		USING (SELECT %s FROM %s s %s WHERE %s) AS s
		   ON 1 = 0
		 WHEN NOT MATCHED THEN
		      INSERT (%s) VALUES (%s)
		%s;`,
		step.Table, strings.Join(selectCols, ", "), step.Table, strings.Join(joins, " "),
		strings.Join(matched, " OR "), joinIdents(cols), strings.Join(values, ", "), output)

	if err := tx.Exec(sql).Error; err != nil {
		return fmt.Errorf("copying %s: %w", step.Table, err)
	}
	return nil
}

// copyClonedProjectHistory gives the new copy the old copy's Change History - header rows to
// the new quotation, tab rows to their tab's copy - in the order it was written. COPIED: the
// old copy stays viewable from the version list with its history as it was.
//
// A history row is filed under either the quotation or one tab's item set, and the two kinds of
// id share one column and can collide, so rows are told apart by label as well as id - the same
// rule the apps use (ProjectHistoryRows).
func copyClonedProjectHistory(tx *gorm.DB, sourceID, newID uint) error {
	const tabLabel = `(h.old_data LIKE 'ITEM/SET %' OR h.old_data LIKE 'Sub Project: %'
	                   OR h.old_data LIKE 'Advance Conditions: %' OR h.old_data LIKE 'Project Items Table -> %'
	                   OR h.old_data LIKE 'Project Wiring Table -> %')`
	const headerLabel = `(h.old_data LIKE 'PROJECT - %' OR h.old_data LIKE 'Header -> %'
	                      OR h.old_data LIKE 'Footer -> %' OR h.old_data LIKE 'Project Header: %'
	                      OR h.old_data LIKE 'Multiplier Table -> %')`

	if err := tx.Exec(`
		INSERT INTO tbl_trans_sales_project_history (based_id, [user], [date], [time], old_data, new_data)
		SELECT x.based_id, x.[user], x.[date], x.[time], x.old_data, x.new_data
		  FROM (
		        SELECT h.history_id, ? AS based_id, h.[user], h.[date], h.[time], h.old_data, h.new_data
		          FROM tbl_trans_sales_project_history h
		         WHERE h.based_id = ? AND NOT `+tabLabel+`
		        UNION ALL
		        SELECT h.history_id, m.new_id, h.[user], h.[date], h.[time], h.old_data, h.new_data
		          FROM tbl_trans_sales_project_history h
		          JOIN #project_clone_map m ON m.kind = 'item_set' AND m.old_id = h.based_id
		         WHERE NOT `+headerLabel+`
		       ) x
		 ORDER BY x.history_id`, newID, sourceID).Error; err != nil {
		return fmt.Errorf("carrying the change history over to the new copy: %w", err)
	}
	return nil
}

// writableColumns lists a table's columns except its identity and computed ones - every
// column a copy must carry, read from the database rather than from a Go model.
func writableColumns(tx *gorm.DB, table string) ([]string, error) {
	var cols []string
	if err := tx.Raw(`
		SELECT c.name FROM sys.columns c
		 WHERE c.object_id = OBJECT_ID(?) AND c.is_identity = 0 AND c.is_computed = 0
		   AND TYPE_NAME(c.system_type_id) NOT IN ('timestamp', 'rowversion')
		 ORDER BY c.column_id`, table).Scan(&cols).Error; err != nil {
		return nil, fmt.Errorf("reading the columns of %s: %w", table, err)
	}
	if len(cols) == 0 {
		return nil, fmt.Errorf("table %s not found", table)
	}
	return cols, nil
}

func quoteIdent(name string) string {
	return "[" + strings.ReplaceAll(name, "]", "]]") + "]"
}

func joinIdents(cols []string) string {
	quoted := make([]string, len(cols))
	for i, c := range cols {
		quoted[i] = quoteIdent(c)
	}
	return strings.Join(quoted, ", ")
}

// InvalidateProjectCopyCaches drops the cached lists that a copy made in raw SQL changed
// behind the model layer's back.
func InvalidateProjectCopyCaches() {
	for _, m := range []interface{}{
		&models.SalesQuotation{}, &models.SalesProjectMultiplier{}, &models.SalesProjectItemSet{},
		&models.SalesProjectContent{}, &models.SalesProjectContentFinal{}, &models.SalesProjectSizeUp{},
		&models.SalesProjectAdvancedConditions{}, &models.SalesProjectItems{},
		&models.SalesQuotationSelectedImage{}, &models.SalesProjectWiring{}, &models.WiringUserInput{},
		&models.SalesProjectHistory{},
	} {
		if err := services.InvalidateCacheByModel(m); err != nil {
			fmt.Println("cache invalidation after project copy:", err)
		}
	}
	_ = invalidateLatestQuotationCache()
	InvalidateEngineeringQuotationListCache()
}
