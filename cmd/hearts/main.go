package main

import (
	"context"
	"heartbeat/client"
	"heartbeat/server"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	cfg, err := client.ReturnConfig("/etc/heartbeat/config.yml")
	if err != nil {
		log.Printf("Couldn't read config file")
		os.Exit(2)
	}

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

	if err := run(ctx, os.Stdout, cfg); err != nil {
		os.Exit(1)
	}
}

func run(ctx context.Context, out io.Writer, cfg client.Config) error {
	log.SetOutput(out)

	if cfg.Server.Enabled {
		err := server.Start(cfg)
		if err != nil {
			return err
		}
		return nil
	} else {
		for {
			select {
			case <-ctx.Done():
				return nil
			case <-time.Tick(240 * time.Second):
				log.Printf("Running data collection")
				data, err := client.CollectionRun()
				if err != nil {
					return err
				}
				sendDataErr := client.SendData(cfg.Database.URI, data)
				if sendDataErr != nil {
					return sendDataErr
				}
			}
		}
	}
}
