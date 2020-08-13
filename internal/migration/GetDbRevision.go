package migration

import (
    "database/sql"
    "errors"
    "github.com/sedl/docsis-pnm/internal/misc"
)

var SchemaVersionError = errors.New("error: multiple rows found in schema_version. Can't determine database revision. Manual intervention needed")

// GetDbRevision returns the database revision. Returns -1 if no schema exists
func GetDbRevision(conn *sql.DB) (int, error) {
    var schemaExists int

    rows, err := conn.Query("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'cmts'")
    if err != nil {
        return 0, err
    }

    rows.Next()
    err = rows.Scan(&schemaExists)
    misc.CloseOrLog(rows)
    if err != nil {
        return 0, err
    }
    if schemaExists == 0 {
        return -1, nil
    }

    rows, err = conn.Query("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'schema_version'")
    if err != nil {
        return 0, err
    }

    rows.Next()
    err = rows.Scan(&schemaExists)
    misc.CloseOrLog(rows)
    if err != nil {
        return 0, err
    }

    // No schema_version table found -> schema revision 0
    if schemaExists == 0 {
        return 0, nil
    }

    rows, err = conn.Query("SELECT COUNT(version) FROM schema_version")
    if err != nil {
        return 0, nil
    }

    var rowCount int
    rows.Next()
    err = rows.Scan(&rowCount)
    misc.CloseOrLog(rows)
    if rowCount > 1 {
        return 0, SchemaVersionError
    }
    if rowCount == 0 {
        return 0, nil
    }

    var schemaVersion int
    rows, err = conn.Query("SELECT version FROM schema_version")
    if err != nil {
        return 0, nil
    }

    rows.Next()
    err = rows.Scan(&schemaVersion)
    if err != nil {
        return 0, err
    }
    misc.CloseOrLog(rows)
    return schemaVersion, nil
}
