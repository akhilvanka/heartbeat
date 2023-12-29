package main

import (
	"context"
	"fmt"
	"heartbeat/client"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var (
	isServer bool // Makeshift variable name for operations now
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	isServer = false

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
					} else {
						log.Printf("Domain %s is down", domain)
					}
				}
			}
		}
	}
}