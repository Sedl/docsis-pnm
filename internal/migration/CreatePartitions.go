package migration

import (
    "database/sql"
    "fmt"
    "log"
)

func CreatePartitions(dbc *sql.DB, partStart int64, interval int64) error {

    for _, part := range partitions {
        query := fmt.Sprintf(part, partStart, partStart, partStart+interval-1)
        log.Println(query)
        _, err := dbc.Exec(query)
        if err != nil {
            return err
        }
    }

    return nil
}
