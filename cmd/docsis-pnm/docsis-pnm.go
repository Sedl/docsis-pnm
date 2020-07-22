package main

import (
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/api"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/manager"
	"log"
	"net/http"
	"os"
)

func errorAndExit(err error) {
	_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
	os.Exit(1)
}

func main() {

	log.Println("debug: loading configuration")
	cfg := config.Read()
	log.Println("debug: config read successful")
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

	router := mux.NewRouter().StrictSlash(true)
	api.Register(router, cmtsManager)
	log.Fatal(http.ListenAndServe(":8080", router))
}

