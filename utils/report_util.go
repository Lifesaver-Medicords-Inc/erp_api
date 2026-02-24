package utils

import (
	"fmt"

	"gorm.io/gorm"
)

func DocNoGenerator(id uint) string {
	DocNo := fmt.Sprintf( /**"RR#" + **/ "%04d", id)
	return DocNo
}

// NextDocNo generates the next document number
func NextDocNo(tx *gorm.DB, model any, column string) (int, error) {
	var last int

	// Safely parse model to get table name
	stmt := &gorm.Statement{DB: tx}
	if err := stmt.Parse(model); err != nil {
		return 0, err
	}

	tableName := stmt.Schema.Table

	query := fmt.Sprintf(`
		SELECT COALESCE(MAX(%s), 0)
		FROM %s WITH (UPDLOCK, HOLDLOCK)
	`, column, tableName)

	if err := tx.Raw(query).Scan(&last).Error; err != nil {
		return 0, err
	}

	return last + 1, nil
}
