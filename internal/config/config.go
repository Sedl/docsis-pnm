package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Snmp    Snmp
	Db      Db
}

func Read() *Config {

	cfg := &Config{}

	v := viper.New()
	v.SetConfigName("docsis-pnm")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/docsis-pnm")

	v.SetDefault("snmp.community", "public")
	v.SetDefault("snmp.retries", 3)
	v.SetDefault("snmp.timeout", 5)

	v.SetDefault("snmp.workercount", 200)
	v.SetDefault("snmp.modempollinterval", 900)
	v.SetDefault("db.commitinterval", 60)

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := v.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return cfg
}

