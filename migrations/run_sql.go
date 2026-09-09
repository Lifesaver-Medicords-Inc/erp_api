package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pierceperado/smpc/initializers"
)

// splitSQLBatches splits a SQL script on standalone "GO" lines, the same way
// sqlcmd/SSMS treat them as batch separators. SQL Server itself has no
// concept of "GO" - it's purely a client-side convention - so statements
// like CREATE/ALTER VIEW/PROC/FUNCTION/TRIGGER that must be the sole
// statement in a batch need to be split out this way before being sent
// over a single Exec call.
var goBatchSeparator = regexp.MustCompile(`(?im)^\s*GO\s*$`)

func splitSQLBatches(script string) []string {
	parts := goBatchSeparator.Split(script, -1)
	batches := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			batches = append(batches, trimmed)
		}
	}
	return batches
}

// runSQLFile executes one script, batch by batch. Returns the first error
// rather than exiting, so the caller can decide whether it is fatal or merely
// premature - see runSQLFolder.
func runSQLFile(file string) error {
	sqlBytes, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("failed reading %s: %w", file, err)
	}

	for _, batch := range splitSQLBatches(string(sqlBytes)) {
		if err := initializers.DB.Exec(batch).Error; err != nil {
			return err
		}
	}

	return nil
}

// runSQLFolder executes every script in a folder, repeating while it keeps
// making progress.
//
// Scripts were previously run once each in glob (alphabetical) order, which is
// only correct if no object depends on one defined later in the alphabet. It
// does: 14 views select from other views, and unlike stored procedures, views
// have no deferred name resolution - the referenced object must exist at ALTER
// time. On a fresh database sql/views/GetBpiSuppliers.sql therefore died on
// "Invalid object name 'vw_items'" because vw_items.sql sorts far later, and
// RunSQLMigrations log.Fatal'd there, taking the API down on its first start
// against a new database - 7 of 84 views created.
//
// Rather than encoding the dependency order in filenames (a numeric prefix per
// file, correct only until someone adds a view and forgets), each pass retries
// whatever failed. A view whose dependency was created by an earlier file in the
// same pass succeeds on the next one. Ordering becomes irrelevant, including for
// views added later.
//
// Termination is on progress, not a fixed retry count: if a whole pass fixes
// nothing, the remaining failures are real - a genuine SQL error, a missing
// table, or a circular dependency - and are reported together rather than one at
// a time, so a fresh-database bring-up shows every problem at once.
func runSQLFolder(path string) {
	pending, err := filepath.Glob(path + "/*.sql")
	if err != nil {
		log.Fatal(err)
	}

	lastErr := make(map[string]error, len(pending))

	for len(pending) > 0 {
		var failed []string

		for _, file := range pending {
			if err := runSQLFile(file); err != nil {
				lastErr[file] = err
				failed = append(failed, file)
				continue
			}
			delete(lastErr, file)
			fmt.Println("Executed:", file)
		}

		// A pass that fixed nothing will not fix anything on the next one.
		if len(failed) == len(pending) {
			var b strings.Builder
			fmt.Fprintf(&b, "SQL execution failed for %d file(s) in %s:", len(failed), path)
			for _, file := range failed {
				fmt.Fprintf(&b, "\n  %s: %v", file, lastErr[file])
			}
			log.Fatal(b.String())
		}

		pending = failed
	}
}

func RunSQLMigrations() {
	runSQLFolder("sql/views")
	runSQLFolder("sql/procedures")
	// Triggers last: tr_inv_item_stocks_ledger depends on tbl_inv_stock_transactions,
	// which initializers.MigrateModel("inventory") creates earlier in main.go's init().
	runSQLFolder("sql/triggers")
}
