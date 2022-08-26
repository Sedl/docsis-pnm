package migration

import "database/sql"

func rev6(db *sql.DB) error {
	query := "ALTER TABLE cmts ADD COLUMN IF NOT EXISTS snmp_max_repetitions INTEGER NOT NULL DEFAULT 0;\n" +
		"UPDATE schema_version SET version = 6;"

	_, err := db.Exec(query)
	return err
}
