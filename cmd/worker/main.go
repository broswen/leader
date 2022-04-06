package main

import (
	"flag"
	"github.com/broswen/leader/pkg/worker"
	"github.com/google/uuid"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	// parse leader addr:port to send health updates
	// create random id? start with set id?
	// store initial config
	// start goroutine to send health update and receive/update local config
	// if selected, start work goroutine
	// if not selected, stop running goroutine
	var leaderAddr string
	flag.StringVar(&leaderAddr, "leader", "http://localhost:8080", "set leader address")

	var port string
	flag.StringVar(&port, "port", ":8080", "set api port")

	var id string
	flag.StringVar(&id, "id", "", "set working id")

	flag.Parse()

	if id == "" {
		uid, err := uuid.NewUUID()
		if err != nil {
			log.Fatal(err)
		}
		id = uid.String()
	}

	worker, err := worker.New(id, leaderAddr)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := http.ListenAndServe(port, worker.Router()); err != nil {
			log.Fatal(err)
		}
	}()

	updateErrors := worker.SendUpdates()

	worker.Work()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-updateErrors:
			log.Printf("updating: %s", err.Error())
		case sig := <-sigChan:
			log.Printf("received signal: %s", sig)
			os.Exit(0)
		}
	}

}
