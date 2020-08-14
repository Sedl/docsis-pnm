package migration

import (
    "database/sql"
    "fmt"
    "log"
    "regexp"
    "strconv"
    "time"
)

var partitionRegex = regexp.MustCompile("(.*)_([0-9]+)$")

// DropOldPartitions drops all partitioned tables that are older than PartitionInterval * PartitionRetentionCount
func DropOldPartitions(conn *sql.DB) error {
    rows, err := conn.Query("SELECT table_name FROM information_schema.tables WHERE table_schema = 'public'")
    if err != nil {
        return err
    }

    dropOlderThan := calculatePartitionStart(time.Now().Unix()) - (PartitionInterval * PartitionRetentionCount)

    for rows.Next() {
        var table string
        err = rows.Scan(&table)
        if err != nil {
            return err
        }

        match := partitionRegex.FindStringSubmatch(table)
        if len(match) == 0 {
            continue
        }

        partitionStart, _ := strconv.Atoi(match[2])
        if int64(partitionStart) < dropOlderThan {
            log.Printf("dropping table %s\n", table)
            _, err := conn.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table))
            if err != nil {
                return err
            }
        }
    }

    return nil
}
