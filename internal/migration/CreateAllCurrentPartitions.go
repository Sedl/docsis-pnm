package migration

import (
    "database/sql"
    "time"
)

func calculatePartitionStart(timestamp int64) int64 {
    return (timestamp / PartitionInterval) * PartitionInterval
}

// CreateAllCurrentPartitions create partitions for the current time plus one in advance
func CreateAllCurrentPartitions(conn *sql.DB) error {

    partStart := calculatePartitionStart(time.Now().Unix())

    if err := CreatePartitions(conn, partStart, PartitionInterval); err != nil {
        return err
    }
    if err := CreatePartitions(conn, partStart+PartitionInterval, PartitionInterval); err != nil {
        return err
    }

    return nil
}
