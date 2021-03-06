package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

func makeDirectories() {
	log.Trace().Msgf("establishing log directory presence at %s...", config.LogDir)
	err := os.MkdirAll(config.LogDir, 0o740)
	if err != nil {
		log.Fatal().
			Str("directory", config.LogDir).Caller().
			Err(err).Msg("failed to open log directory")
		return
	}

	log.Trace().Msgf("establishing data directory presence at %s...", config.DBDir)
	err = os.MkdirAll(config.DBDir, 0o740)
	if err != nil {
		log.Fatal().
			Str("directory", config.DBDir).Caller().
			Err(err).Msg("failed to open data directory")
		return
	}
}

func wait(hs *http.Server) {
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	for {
		select {
		case <-c:
			log.Warn().Msg("Interrupt detected, shutting down gracefully...")
			if err := hs.Shutdown(ctx); err != nil {
				cancel()
			}
			log.Print("fin.")
			cancel()
			return
		}
	}
}

func main() {
	config.Init()
	config.PrintBanner()
	makeDirectories()
	log.Debug().Msg("debug enabled")
	log.Trace().Msg("trace enabled")
	err := dbInit()
	if err != nil {
		log.Fatal().Err(err).Msg("bitcask failure")
	}
	defer func() {
		err := db.SyncAndCloseAll()
		if err != nil {
			log.Warn().Err(err).Msg("sync failure!")
		}
	}()
	go func() {
		err := serveTermbin()
		if err != nil {
			log.Fatal().Err(err).Msg("failed to start termbin")
		}
	}()
	time.Sleep(50 * time.Millisecond)
	wait(httpRouter())
}
