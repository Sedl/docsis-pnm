package cmd

import (
	"context"
	"fmt"
	"github.com/sedl/docsis-pnm/internal/api"
	"github.com/sedl/docsis-pnm/internal/manager"
	"github.com/sedl/docsis-pnm/internal/types"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

// Run starts data collection from CMTS and modems
func Run(cfg *types.Config) {
	errs := cfg.Validate()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Printf("CONFIG_ERROR: %s", err.Error())
		}
		os.Exit(ErrConfig)
	}

	cmtsManager, err := manager.NewManager(cfg)
	if err != nil {
		errorAndExit(err)
	}

	err = cmtsManager.Run()
	if err != nil {
		errorAndExit(err)
	}

	err = cmtsManager.AddAllCmtsFromDb()
	if err != nil {
		errorAndExit(err)
	}

	log.Println("debug: init done")

	server := api.NewApi(cmtsManager, &cfg.Api)
	wg := registerExitHandler(cmtsManager, server)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()
}

func errorAndExit(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

// registerExitHandler is responsible for a safe shutdown of the application
func registerExitHandler(manager *manager.Manager, server *http.Server) *sync.WaitGroup {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		<-c
		defer wg.Done()
		manager.Stop()
		err := server.Shutdown(context.Background())
		if err != nil {
			log.Printf("error: http shutdown failed: %s\n", err.Error())
		}
	}()

	return wg

}
