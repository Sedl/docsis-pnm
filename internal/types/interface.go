package types

type DbInterface interface {
//	CreateTables() error
// 	CreatePartitionTables() error
//	InitDb() error

	GetCMTSByHostname(hostname string) (*CMTSRecord, error)
	GetCMTSUpstreamByDescr(cmtsDbID uint32, description string) (*CMTSUpstreamRecord, error)
	UpdateCmtsUpstreams (records map[int]*CMTSUpstreamRecord) error
	InsertCMTSUpstreamHistory(record *CMTSUpstreamHistoryRecord) error
	UpdateModemFromModemInfo(minfo *ModemInfo) error
}

type ModemUpdaterInterface interface {
	UpdateModemData(data *ModemData) error
}

/*
type CmtsUpdaterInterface interface {
	UpdateCmtsModemInfo(info *ModemInfo) error
}
 */

type ModemPollWorkerInterface interface {
	Poll(request *ModemPollRequest) error
}