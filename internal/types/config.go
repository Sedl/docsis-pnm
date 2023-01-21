package types

import (
	"errors"
)

type ApiConfig struct {
	ListenAddress string `yaml:"listenaddress"`
}

type Config struct {
	Snmp  Snmp      `yaml:"snmp"`
	Db    Db        `yaml:"db"`
	Tftp  Tftp      `yaml:"tftp"`
	Api   ApiConfig `yaml:"api"`
	Debug bool      `yaml:"debug"`
}

func (c *Config) Validate() []error {
	var errs = make([]error, 0)

	// SNMP values
	if c.Snmp.Retries < 0 {
		errs = append(errs, errors.New("snmp.retries has to be a positive integer"))
	}

	if c.Snmp.Community == "" {
		errs = append(errs, errors.New("snmp.community can not be an empty string"))
	}

	if c.Snmp.Timeout < 1 {
		errs = append(errs, errors.New("snmp.timeout can not be lower than 1"))
	}

	if c.Snmp.ModemPollInterval < 0 {
		errs = append(errs, errors.New("snmp.modempollinterval has to be a positive integer"))
	}

	if c.Snmp.WorkerCount < 1 {
		errs = append(errs, errors.New("snmp.workercount can't be lower than 1"))
	}

	// Db values

	if c.Db.Connstr == "" {
		errs = append(errs, errors.New("db.connstr can not be empty"))
	}

	if c.Db.CommitInterval < 1 {
		errs = append(errs, errors.New("db.commitinterval can not be lower than 1"))
	}

	return errs
}

type Snmp struct {
	Community         string `yaml:"community"`
	Timeout           int    `yaml:"timeout"`
	Retries           int    `yaml:"retries"`
	WorkerCount       int    `yaml:"workercount"`
	ModemPollInterval int    `yaml:"modempollinterval"`
}

type Db struct {
	Connstr        string `yaml:"connstr"`
	CommitInterval int64  `yaml:"commitinterval"`
}

type Tftp struct {
	ExternalAddress string `json:"externaladdress"`
}
