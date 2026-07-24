package migrations

import (
	"fmt"

	"gorm.io/gorm"
)

// execAll runs statements in order, stopping at the first failure.
func execAll(tx *gorm.DB, statements ...string) error {
	for _, stmt := range statements {
		if err := tx.Exec(stmt).Error; err != nil {
			return fmt.Errorf("migrations: executing %q: %w", truncate(stmt, 120), err)
		}
	}
	return nil
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
