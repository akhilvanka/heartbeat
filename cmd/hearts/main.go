package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"heartbeat/client"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	isServer  bool // Makeshift variable name for operations now
	serverURL string
)

type Payload struct {
	SystemStatus int
	Services     []Domain
}

type Domain struct {
	ServiceName   string
	ServiceStatus int
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	isServer = false
	serverURL = "http://lcoalhost:3000"

	signalChange := make(chan os.Signal, 1)
	signal.Notify(signalChange, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)

	defer func() {
		signal.Stop(signalChange)
		cancel()
	}()

	go func() {
		select {
		case s := <-signalChange:
			switch s {
			case syscall.SIGINT, syscall.SIGTERM:
				log.Printf("Got SIGINT/SIGTERM, exiting.")
				cancel()
				os.Exit(1)
			case syscall.SIGHUP:
				log.Printf("Got SIGHUP, reloading.")
			}
		case <-ctx.Done():
			log.Printf("Done.")
			os.Exit(1)
		}
	}()

	if err := run(ctx, os.Stdout); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer) error {

	log.SetOutput(out)
	var statusList []Domain

	if isServer {
		return nil
	} else {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.Tick(1 * time.Second):
				data, err := client.ReadCaddyfile()
				if err != nil {
					return err
				}
				for _, domain := range data {
					status := client.QueryURL(fmt.Sprintf("https://%s", domain))
					if status {
						log.Printf("Domain %s is up", domain)
						a := Domain{
							ServiceName:   domain,
							ServiceStatus: 1,
						}
						statusList = append(statusList, a)
					} else {
						log.Printf("Domain %s is down", domain)
						a := Domain{
							ServiceName:   domain,
							ServiceStatus: 0,
						}
						statusList = append(statusList, a)
					}
				}
				payloadMessage := Payload{
					SystemStatus: 1,
					Services:     statusList,
				}
				body, _ := json.Marshal(payloadMessage)

				// Send JSON status
				resp, err := http.Post(serverURL, "application/json", bytes.NewBuffer(body))
				if err != nil {
					return err
				}
				defer func(Body io.ReadCloser) {
					err := Body.Close()
					if err != nil {
						log.Printf("Error found when closing body")
					}
				}(resp.Body) // Handle this and check if it actually leaks (Shouldn't but just to be safe)
				if resp.StatusCode != 200 {
					return errors.New("unable to post data to server")
				}
			}
		}
	}
}
