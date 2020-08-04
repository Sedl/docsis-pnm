package db

import "database/sql"

func rev2(db *sql.DB) error {
    query := `CREATE TABLE IF NOT EXISTS cmts_upstream_history_modem (
    modem_id INTEGER,
    us_id INTEGER,
    poll_time INTEGER,
    power_rx INTEGER,
    snr INTEGER,
    microrefl INTEGER,
    unerroreds BIGINT,
    correcteds BIGINT,
    erroreds BIGINT
) PARTITION BY RANGE (poll_time);
CREATE UNIQUE INDEX IF NOT EXISTS idx_cmts_mdmus ON cmts_upstream_history_modem(modem_id, poll_time, us_id);
CREATE TABLE IF NOT EXISTS cmts_upstream_history_modem_def PARTITION OF cmts_upstream_history_modem DEFAULT;
DROP TABLE IF EXISTS modem_upstream_cmts;

UPDATE schema_version SET version = 2;
`
    _, err := db.Exec(query)
    return err
}