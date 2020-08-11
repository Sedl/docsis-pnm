package db

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"time"
)

const (
	// PartitionInterval defines the interval in seconds by which the performace data tables are getting partitioned
	PartitionInterval = 86400
)

var partitions = []string{
	"CREATE TABLE IF NOT EXISTS modem_downstream_%d PARTITION OF modem_downstream FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_upstream_%d PARTITION OF modem_upstream FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_data_%d PARTITION OF modem_data FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS cmts_upstream_history_%d PARTITION OF cmts_upstream_history FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_ofdm_downstream_%d PARTITION OF modem_ofdm_downstream FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS cmts_upstream_history_modem_%d PARTITION OF cmts_upstream_history_modem FOR VALUES FROM (%d) TO (%d)",
}

type revUpgradeFunc = func (db *sql.DB) error


var revisions = []revUpgradeFunc{
	rev0,
	rev1,
	rev2,
	rev3,
	rev4,
}

func createPartitions(dbc *sql.DB, partStart int64, interval int64) error {

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
// CreateDB creates the database structure and partition tables for the current timespan
func (db *Postgres) InitDb() error {
	if err := db.CreateTables(); err != nil {
		return err
	}
	if err := db.CreatePartitionTables(); err != nil {
		return err
	}
	return nil
}

// CreatePartitionTables creates the partition on the database backend
func (db *Postgres) CreatePartitionTables() error {
	dbc, err := db.GetConn()
	if err != nil {
		return err
	}

	partStart := (time.Now().Unix() / PartitionInterval) * PartitionInterval

	if err = createPartitions(dbc, partStart, PartitionInterval); err != nil {
		return err
	}
	if err = createPartitions(dbc, partStart+PartitionInterval, PartitionInterval); err != nil {
		return err
	}

	return nil
}

// GoMaintenanceWorker is a background task to automatically create database partition tables
// and should handle other database maintenance stuff like expunge of old records after the expiration time
func (db *Postgres) GoMaintenanceWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			if err := db.CreatePartitionTables(); err != nil {
				log.Printf("Error in goroutine GoMaintenanceWorker while creating partition tables: %s\n", err)
				continue
			}
		}
	}
}

func (db *Postgres) CreateTables() error {
	conn, err := db.GetConn()
	if err != nil {
		return err
	}
	/*
	_, err = conn.Exec(Tables)
	if err != nil {
		return err
	}
	 */
	version, err := db.getDbRevision()
	if err != nil {
		return err
	}

	maxRevision := len(revisions) - 1
	log.Printf("debug: current db revision %d, upgrading to %d\n", version, maxRevision)

	for i := version+1; i <= maxRevision; i++ {
		log.Printf("debug: Upgrading database schema to version %d\n", i)
		err = revisions[i](conn)
		if err != nil {
			return err
		}
	}

	return nil
}

var schemaVersionError = errors.New("error: multiple rows found in schema_version. Can't determine database revision. Manual intervention needed")

// getDbRevision returns the database revision. Returns -1 if no schema exists
func (db *Postgres) getDbRevision() (int, error) {
	conn, err := db.GetConn()
	if err != nil {
		return 0, err
	}

	var schemaExists int

	rows, err := conn.Query("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'cmts'")
	if err != nil {
		return 0, err
	}

	rows.Next()
	err = rows.Scan(&schemaExists)
	CloseOrLog(rows)
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
	CloseOrLog(rows)
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
	CloseOrLog(rows)
	if rowCount > 1 {
		return 0, schemaVersionError
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
	CloseOrLog(rows)
	return schemaVersion, nil
}