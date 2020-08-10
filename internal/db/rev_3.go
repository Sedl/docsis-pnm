package db

import "database/sql"

func rev3(db *sql.DB) error {
    query := "ALTER TABLE cmts_upstream ALTER COLUMN descr SET NOT NULL;\n" +
        "ALTER TABLE cmts_upstream ALTER COLUMN alias SET NOT NULL;\n" +
        "UPDATE schema_version SET version = 3;"

    _, err := db.Exec(query)
    return err
}
