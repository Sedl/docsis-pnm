package cmd

import (
	"fmt"
	"github.com/sedl/docsis-pnm/internal/logger"
	"github.com/sedl/docsis-pnm/internal/types"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"strings"
)

var cfgFile string
var rootCmd = &cobra.Command{
	Use:   "docsis-pnm",
	Short: "Collects performance data from DOCSIS CMTS and modems",
}

func CobraExecute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(ErrCobra)
	}
}
func CobraInit(cfg *types.Config) {
	cobra.OnInitialize(func() { configInit(cfg) })

	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/docsis-pnm.toml or ./docsis-pnm.toml)")

	rootCmd.PersistentFlags().String("snmp.community", "", "Default SNMP community for modems and CMTS (default: public)")
	viper.BindPFlag("snmp.community", rootCmd.PersistentFlags().Lookup("snmp.community"))
	viper.SetDefault("snmp.community", "public")

	rootCmd.PersistentFlags().Int("snmp.timeout", 0, "SNMP timeout in seconds for modems and CMTS (default: 5)")
	viper.BindPFlag("snmp.timeout", rootCmd.PersistentFlags().Lookup("snmp.timeout"))
	viper.SetDefault("snmp.timeout", 5)

	rootCmd.PersistentFlags().Int("snmp.retries", 0, "Retries if a SNMP request failed (default: 3)")
	viper.BindPFlag("snmp.retries", rootCmd.PersistentFlags().Lookup("snmp.retries"))
	viper.SetDefault("snmp.retries", 3)

	rootCmd.PersistentFlags().Int("snmp.workercount", 0, "Number of SNMP modem poll workers (default: 200)")
	viper.BindPFlag("snmp.workercount", rootCmd.PersistentFlags().Lookup("snmp.workercount"))
	viper.SetDefault("snmp.workercount", 200)

	rootCmd.PersistentFlags().Int("snmp.modempollinterval", 0, "Interval in which the modems get polled (default: 900)")
	viper.BindPFlag("snmp.modempollinterval", rootCmd.PersistentFlags().Lookup("snmp.modempollinterval"))
	viper.SetDefault("snmp.modempollinterval", 900)

	rootCmd.PersistentFlags().String("db.connstr", "", "Database connection string")
	viper.BindPFlag("db.connstr", rootCmd.PersistentFlags().Lookup("db.connstr"))
	viper.SetDefault("db.connstr", "")

	rootCmd.PersistentFlags().Int("db.commitinterval", 0, "Interval in which a database \"COMMIT\" is issued when bulk inserting new modem data (default: 60)")
	viper.BindPFlag("db.commitinterval", rootCmd.PersistentFlags().Lookup("db.commitinterval"))
	viper.SetDefault("db.commitinterval", 60)

	rootCmd.PersistentFlags().String("tftp.externaladdress", "", "(Own) TFTP address where modems can submit performance data to. Required for i.E. DOCSIS 3.1 OFDM MER performance data. Default is disabled.")
	viper.BindPFlag("tftp.externaladdress", rootCmd.PersistentFlags().Lookup("tftp.externaladdress"))
	viper.SetDefault("tftp.externaladdress", "")

	rootCmd.PersistentFlags().String("api.listenaddress", "", "Set listen IP and port (default: 0.0.0.0:8080)")
	viper.BindPFlag("api.listenaddress", rootCmd.PersistentFlags().Lookup("api.listenaddress"))
	viper.SetDefault("api.listenaddress", "0.0.0.0:8080")

	rootCmd.PersistentFlags().Bool("debug", false, "Enable debug mode (verbose logging)")
	viper.BindPFlag("debug", rootCmd.PersistentFlags().Lookup("debug"))
	viper.SetDefault("debug", false)

	var runCmd = &cobra.Command{
		Use:   "run",
		Short: "Run data collection",
		Run: func(cmd *cobra.Command, args []string) {
			Run(cfg)
		},
	}
	rootCmd.AddCommand(runCmd)

	var configPrintCmd = &cobra.Command{
		Use:   "config-print",
		Short: "Prints the configuration. For validation see `config-validate`",
		Run: func(_ *cobra.Command, _ []string) {
			ConfigPrint(cfg)
		},
	}
	rootCmd.AddCommand(configPrintCmd)

	var configValidateCmd = &cobra.Command{
		Use:   "config-validate",
		Short: "Validates the configuration",
		Run: func(_ *cobra.Command, _ []string) {
			ConfigValidate(cfg)
		},
	}
	rootCmd.AddCommand(configValidateCmd)
}

func configInit(cfg *types.Config) {
	// get configuration file
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		cobra.CheckErr(err)
		viper.AddConfigPath(home)
		viper.AddConfigPath(".")
		viper.SetConfigName("docsis-pnm")

	}

	viper.SetEnvPrefix("pnm")
	viper.AutomaticEnv()

	// read configuration file
	if err := viper.ReadInConfig(); err == nil {
		logger.Info(fmt.Sprint("Using config file: ", viper.ConfigFileUsed()))
	}

	viper.Unmarshal(cfg)
}
