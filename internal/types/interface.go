package types

type DbInterface interface {
//	CreateTables() error
// 	CreatePartitionTables() error
//	InitDb() error

	GetCMTSByHostname(hostname string) (*CMTSRecord, error)
	GetCMTSUpstreamByDescr(cmtsDbID int32, description string) (*CMTSUpstreamRecord, error)
	UpdateCmtsUpstreams (records map[int]*CMTSUpstreamRecord) error
	InsertCMTSUpstreamHistory(record *CMTSUpstreamHistoryRecord) error
	UpdateModemFromModemInfo(minfo *ModemInfo) error
}

type ModemUpdaterInterface interface {
	UpdateModemData(data *ModemData) error
}

type ModemPollWorkerInterface interface {
	Poll(request *ModemPollRequest) error
}