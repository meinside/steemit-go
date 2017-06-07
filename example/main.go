package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/meinside/steemit-go"

	"github.com/meinside/steemit-go/api/login"
)

func main() {
	log.Println("Application starting...")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	client, err := steemit.NewClient(true)
	if err != nil {
		panic(err)
	}

	done := make(chan struct{})

	go func(client *steemit.Client) {
		for {
			select {
			case <-interrupt:
				log.Println("Closing connection...")

				if err := client.Close(); err != nil {
					log.Printf("Failed to close connection: %s", err)
				}

				done <- struct{}{}

				return
			}
		}
	}(client)

	// login_api
	if success, err := login.Login(client, "meinside", "XXXXXXXXXXXXXXXXXXXX"); success {
		log.Printf("login successful")
	} else {
		log.Printf("login error: %s", err)
	}
	if res, err := login.GetApiByName(client, "login_api"); err == nil {
		log.Printf("get_api_by_name: %+v", res)
	} else {
		log.Printf("get_api_by_name error: %s", err)
	}
	if res, err := login.GetVersion(client); err == nil {
		log.Printf("get_version: %+v", res)
	} else {
		log.Printf("get_version error: %s", err)
	}

	// wait...
	<-done

	log.Println("Application stopped")
}
