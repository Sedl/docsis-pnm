package migration

import (
    "database/sql"
    "time"
)

// CreateAllCurrentPartitions create partitions for the current time plus one in advance
func CreateAllCurrentPartitions(conn *sql.DB) error {

    partStart := (time.Now().Unix() / PartitionInterval) * PartitionInterval

    if err := CreatePartitions(conn, partStart, PartitionInterval); err != nil {
        return err
    }
    if err := CreatePartitions(conn, partStart+PartitionInterval, PartitionInterval); err != nil {
        return err
    }

    return nil
}
