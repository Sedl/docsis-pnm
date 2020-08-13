package migration

import "database/sql"

func rev1(db *sql.DB) error {
    query := "CREATE TABLE schema_version (version INT); INSERT INTO schema_version (version) VALUES (1);"
    _, err := db.Exec(query)
    return err
}
