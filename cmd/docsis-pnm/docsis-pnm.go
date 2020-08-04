package main

import (
	"context"
	"fmt"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"github.com/sedl/docsis-pnm/internal/api"
	"github.com/sedl/docsis-pnm/internal/config"
	"github.com/sedl/docsis-pnm/internal/manager"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
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
	server := &http.Server{
		Addr:              ":8080",
		Handler: logRequest(router),
	}

	wg := registerExitHandler(cmtsManager, server)

	go func() {
		err = server.ListenAndServe()
		if err != nil {
			log.Fatal(err)
		}
	}()

	wg.Wait()

}

func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}


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