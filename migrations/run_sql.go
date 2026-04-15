package migrations

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/pierceperado/smpc/initializers"
)

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

		err = initializers.DB.Exec(string(sqlBytes)).Error
		if err != nil {
			log.Fatal("SQL execution failed:", file, err)
		}

		fmt.Println("Executed:", file)
	}
}

func RunSQLMigrations() {
	runSQLFolder("sql/views")
	runSQLFolder("sql/procedures")
}
