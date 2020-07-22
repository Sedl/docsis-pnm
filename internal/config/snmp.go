package config

type Snmp struct {
	Community string
	// TODO implement this
	Timeout int
	// TODO implement this
	Retries int
	WorkerCount int
	ModemPollInterval int
}
