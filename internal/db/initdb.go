package db

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

const (
	// PartitionInterval defines the interval in seconds by which the performace data tables are getting partitioned
	PartitionInterval = 86400
)

const Tables =
` CREATE TABLE IF NOT EXISTS cmts (
  id SERIAL PRIMARY KEY,
  hostname VARCHAR(255) NOT NULL,
  snmp_community VARCHAR(255),
  snmp_community_modem VARCHAR(255),
  disabled BOOLEAN DEFAULT FALSE NOT NULL,
  poll_interval INT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cmts_hostname ON cmts (hostname);


CREATE TABLE IF NOT EXISTS modem (
  id BIGSERIAL PRIMARY KEY,
  mac MACADDR NOT NULL,
  sysdescr TEXT,
  ip INET,
  cmts_id INTEGER REFERENCES cmts(id) NOT NULL,
  snmp_index INTEGER NOT NULL,
  docsis_ver INTEGER NOT NULL,
  ds_primary INT NOT NULL,
  cmts_ds_idx INT NOT NULL,
  cmts_us_idx INT NOT NULL
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_mac ON modem (mac);
CREATE INDEX IF NOT EXISTS idx_modem_ip ON modem(ip);
CREATE INDEX IF NOT EXISTS idx_modem_cmts_id ON modem(cmts_id);


CREATE TABLE IF NOT EXISTS modem_data (
  modem_id BIGINT,
  poll_time BIGINT,
  error_timeout BOOL
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_data ON modem_data (modem_id, poll_time);
CREATE INDEX IF NOT EXISTS idx_modem_data_timeout ON modem_data (error_timeout);
CREATE TABLE IF NOT EXISTS modem_data_def PARTITION OF modem_data DEFAULT;

CREATE TABLE IF NOT EXISTS modem_downstream (
  modem_id INTEGER,
  poll_time BIGINT,
  freq INTEGER,
  power FLOAT,
  snr FLOAT,
  microrefl INTEGER,
  unerroreds BIGINT,
  correcteds BIGINT,
  erroreds BIGINT,
  modulation INTEGER
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_downstream ON modem_downstream (modem_id, poll_time, freq);
CREATE TABLE IF NOT EXISTS modem_downstream_def PARTITION OF modem_downstream DEFAULT;


CREATE TABLE IF NOT EXISTS modem_upstream (
    modem_id INTEGER,
    poll_time BIGINT,
    freq  INTEGER,
    modulation INTEGER,
    timing_offset INTEGER
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_upstream ON modem_upstream (modem_id, poll_time, freq);
CREATE TABLE IF NOT EXISTS modem_upstream_def PARTITION OF modem_upstream DEFAULT;


CREATE TABLE IF NOT EXISTS modem_upstream_cmts (
    modem_id INTEGER,
    poll_time BIGINT,
    upstream_id INTEGER,
    power_rx INTEGER,
    status INTEGER,
    unerroreds BIGINT,
    correcteds BIGINT,
    erroreds BIGINT
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_cmts ON modem_upstream_cmts (modem_id, poll_time);
CREATE TABLE IF NOT EXISTS modem_upstream_cmts_def PARTITION OF modem_upstream_cmts DEFAULT;


CREATE TABLE IF NOT EXISTS cmts_upstream (
    id SERIAL PRIMARY KEY,
    cmts_id INTEGER REFERENCES cmts (id) NOT NULL,
    snmp_idx INTEGER NOT NULL,
    descr VARCHAR,
    freq INTEGER NOT NULL,
    alias VARCHAR,
    admin_status INTEGER NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_cmts_upstream_cmts_id ON cmts_upstream (cmts_id, snmp_idx);
CREATE INDEX IF NOT EXISTS idx_cmts_upstream_descr ON cmts_upstream (descr);

CREATE TABLE IF NOT EXISTS cmts_upstream_history (
    upstream_id INTEGER NOT NULL,
    poll_time BIGINT NOT NULL,
    unerroreds BIGINT NOT NULL,
    correcteds BIGINT NOT NULL,
    erroreds BIGINT NOT NULL,
    utilization INTEGER NOT NULL,
    pkts_broadcast BIGINT NOT NULL,
    pkts_unicast BIGINT NOT NULL,
    bytes BIGINT NOT NULL,
    mer_db FLOAT NOT NULL
) PARTITION BY RANGE(poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cmts_upstream_history ON cmts_upstream_history (upstream_id, poll_time);
CREATE TABLE IF NOT EXISTS cmts_upstream_history_def PARTITION OF cmts_upstream_history DEFAULT;

CREATE TABLE IF NOT EXISTS modem_ofdm_downstream (
    modem_id BIGINT,
    poll_time BIGINT,
    freq INTEGER,
    power FLOAT
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_modem_ofdm_downstream ON modem_ofdm_downstream (modem_id, poll_time);
CREATE TABLE IF NOT EXISTS modem_ofdm_downstream_def PARTITION OF modem_ofdm_downstream DEFAULT;
`

var partitions = []string{
	"CREATE TABLE IF NOT EXISTS modem_downstream_%d PARTITION OF modem_downstream FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_upstream_%d PARTITION OF modem_upstream FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_upstream_cmts_%d PARTITION OF modem_upstream_cmts FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_data_%d PARTITION OF modem_data FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS cmts_upstream_history_%d PARTITION OF cmts_upstream_history FOR VALUES FROM (%d) TO (%d)",
	"CREATE TABLE IF NOT EXISTS modem_ofdm_downstream_%d PARTITION OF modem_ofdm_downstream FOR VALUES FROM (%d) TO (%d)",
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
	_, err = conn.Exec(Tables)
	if err != nil {
		return err
	}
	return nil
}

