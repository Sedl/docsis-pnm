package config

import (
	"github.com/spf13/viper"
	"log"
	"strings"
)

type Config struct {
	Snmp    Snmp
	Db      Db
	Tftp	Tftp
}

func Read() *Config {

	cfg := &Config{}

	v := viper.New()
	replacer := strings.NewReplacer(".", "_")
	v.SetEnvKeyReplacer(replacer)

	v.SetConfigName("docsis-pnm")
	v.AddConfigPath(".")
	v.AddConfigPath("/etc/docsis-pnm")

	v.SetEnvPrefix("pnm")
	v.AutomaticEnv()

	// Don't forget to call SetDefault on every key, or the environment variables won't work
	v.SetDefault("snmp.community", "public")
	v.SetDefault("snmp.timeout", 5)
	v.SetDefault("snmp.retries", 3)
	v.SetDefault("snmp.workercount", 200)
	v.SetDefault("snmp.modempollinterval", 900)

	v.SetDefault("db.connstr", "")
	v.SetDefault("db.commitinterval", 60)

	// set defaults
	v.SetDefault("tftp.externaladdress", "")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}
	err := v.Unmarshal(cfg)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	return cfg
}