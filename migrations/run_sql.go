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

func runSQLFolder(path string) {
	files, err := filepath.Glob(path + "/*.sql")
	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			log.Fatal("Failed reading:", file)
		}

		for _, batch := range splitSQLBatches(string(sqlBytes)) {
			err = initializers.DB.Exec(batch).Error
			if err != nil {
				log.Fatal("SQL execution failed:", file, err)
			}
		}

		fmt.Println("Executed:", file)
	}
}

func RunSQLMigrations() {
	runSQLFolder("sql/views")
	runSQLFolder("sql/procedures")
}
