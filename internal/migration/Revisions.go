package migration

import "database/sql"

type RevUpgradeFunc = func(db *sql.DB) error

var Revisions = []RevUpgradeFunc{
    rev0,
    rev1,
    rev2,
    rev3,
    rev4,
}

var partitions = []string{
    "CREATE TABLE IF NOT EXISTS modem_downstream_%d PARTITION OF modem_downstream FOR VALUES FROM (%d) TO (%d)",
    "CREATE TABLE IF NOT EXISTS modem_upstream_%d PARTITION OF modem_upstream FOR VALUES FROM (%d) TO (%d)",
    "CREATE TABLE IF NOT EXISTS modem_data_%d PARTITION OF modem_data FOR VALUES FROM (%d) TO (%d)",
    "CREATE TABLE IF NOT EXISTS cmts_upstream_history_%d PARTITION OF cmts_upstream_history FOR VALUES FROM (%d) TO (%d)",
    "CREATE TABLE IF NOT EXISTS modem_ofdm_downstream_%d PARTITION OF modem_ofdm_downstream FOR VALUES FROM (%d) TO (%d)",
    "CREATE TABLE IF NOT EXISTS cmts_upstream_history_modem_%d PARTITION OF cmts_upstream_history_modem FOR VALUES FROM (%d) TO (%d)",
}

