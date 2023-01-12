package cmd

import (
    "fmt"
    "github.com/sedl/docsis-pnm/internal/types"
    "gopkg.in/yaml.v3"
    "log"
    "os"
)

func ConfigValidate(cfg *types.Config) {

	errs := cfg.Validate()
	if len(errs) > 0 {
		printConfigErrors(errs)
		os.Exit(ErrConfig)
	} else {
		os.Exit(0)
	}
}

func printConfigErrors(errs []error) {

	for _, err := range errs {
		log.Printf("CONFIG_ERROR: %s", err.Error())
	}

}

func ConfigPrint(cfg *types.Config) {

	yamld, _ := yaml.Marshal(cfg)
	fmt.Println(string(yamld))
}

