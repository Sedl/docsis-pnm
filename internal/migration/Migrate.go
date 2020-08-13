package migration

import (
    "database/sql"
    "log"
)

const (
    // PartitionInterval defines the interval in seconds by which the performace data tables are getting partitioned
    PartitionInterval = 86400
)

// Migrate checks the current database schema revision and updates the schema appropriately
func Migrate(conn *sql.DB) error {

    version, err := GetDbRevision(conn)
    if err != nil {
        return err
    }

    maxRevision := len(Revisions) - 1
    log.Printf("debug: current db revision %d, upgrading to %d\n", version, maxRevision)

    for i := version + 1; i <= maxRevision; i++ {
        log.Printf("debug: Upgrading database schema to version %d\n", i)
        err = Revisions[i](conn)
        if err != nil {
            return err
        }
    }

    return nil
}
