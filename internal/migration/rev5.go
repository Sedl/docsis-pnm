package migration

import "database/sql"

func rev5(db *sql.DB) error {
    query := "ALTER TABLE modem_data ADD COLUMN IF NOT EXISTS bytes_down BIGINT NOT NULL DEFAULT 0;\n" +
        "ALTER TABLE modem_data ADD COLUMN IF NOT EXISTS bytes_up BIGINT NOT NULL DEFAULT 0;\n" +
        "UPDATE schema_version SET version = 5;"

    _, err := db.Exec(query)
    return err
}

