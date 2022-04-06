package main

import (
	"flag"
	"github.com/broswen/leader/pkg/leader"
	"log"
	"net/http"
	"os"
	"os/signal"
)

func main() {
	//parse api port
	var port string
	flag.StringVar(&port, "port", ":8080", "set api port")

	//parse how many wanted workers
	var wantedWorkers int
	flag.IntVar(&wantedWorkers, "workers", 1, "set wanted workers")

	flag.Parse()

	leader, err := leader.New(port, wantedWorkers)
	if err != nil {
		log.Fatal(err)
	}

	//start update cycle
	leader.Update()

	//start api server and listen for updates
	errChan := make(chan error, 1)
	go func() {
		log.Printf("leader listening on %s...", leader.Address)
		if err := http.ListenAndServe(leader.Address, leader.Router()); err != nil {
			errChan <- err
		}
	}()

	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, os.Interrupt, os.Kill)

	for {
		select {
		case err := <-errChan:
			log.Fatal(err)
		case sig := <-sigChan:
			log.Printf("received signal: %s", sig)
			os.Exit(0)
		}
	}
}
