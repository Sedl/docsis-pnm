package manager

type Stats struct {
	QueueLength     int    `json:"queue_length"`
	DbQueueLength   int    `json:"db_queue_length"`
	PollsSuccessful uint64 `json:"polls_successful"`
	PollsErrors     uint64 `json:"polls_errors"`
	ModemsOnline    int32  `json:"modems_online"`
	ModemsOffline   int32  `json:"modems_offline"`
	ActiveCmtsCount int    `json:"active_cmts_count"`
}

func (m *Manager) Stats() Stats {
	stats := Stats{
		QueueLength:     m.modemPoller.GetRequestQueueLength(),
		DbQueueLength:   len(m.modemPoller.ModemDataSink),
		PollsSuccessful: m.modemPoller.GetPollOk(),
		PollsErrors:     m.modemPoller.GetPollErr(),
		ActiveCmtsCount: len(m.cmtsList),
	}

	for _, cmts := range m.cmtsList {
		stats.ModemsOnline += cmts.ValueOfModemsOnline()
		stats.ModemsOffline += cmts.ValueOfModemsOffline()
	}

	return stats
}
