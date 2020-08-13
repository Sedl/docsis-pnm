package migration

import "database/sql"

func rev4(db *sql.DB) error {
    query := "ALTER TABLE modem_upstream ADD COLUMN IF NOT EXISTS tx_power INTEGER NOT NULL DEFAULT 0;\n" +
        "ALTER TABLE modem_upstream DROP COLUMN IF EXISTS modulation;\n" +
        "UPDATE schema_version SET version = 4;"

    _, err := db.Exec(query)
    return err
}
