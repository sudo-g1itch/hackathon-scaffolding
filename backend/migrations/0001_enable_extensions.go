package migrations

import (
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// m0001EnableExtensions turns on the PostgreSQL extensions the schema relies on.
func m0001EnableExtensions() *gormigrate.Migration {
	return &gormigrate.Migration{
		ID: "0001_enable_extensions",
		Migrate: func(tx *gorm.DB) error {
			return execAll(tx,
				`CREATE EXTENSION IF NOT EXISTS "pgcrypto"`,
				`CREATE EXTENSION IF NOT EXISTS "citext"`,
				`CREATE EXTENSION IF NOT EXISTS "uuid-ossp"`,
			)
		},
		Rollback: func(tx *gorm.DB) error {
			return execAll(tx,
				`DROP EXTENSION IF EXISTS "uuid-ossp"`,
				`DROP EXTENSION IF EXISTS "citext"`,
				`DROP EXTENSION IF EXISTS "pgcrypto"`,
			)
		},
	}
}
