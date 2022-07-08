package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/muesli/termenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var Banner string = "CiAgLGQgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogIDg4ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKTU04OE1NTSAsYWRQUFliYSwgOGIsZFBQWWJhLCAgICAgICxhZFBQWVliYSwgICxhZFBQWWJhLCAgCiAgODggICBhOCIgICAgICIiIDg4UCcgICAgIjhhICAgICAiIiAgICAgYFk4IGE4IiAgICAgIiIgIAogIDg4ICAgOGIgICAgICAgICA4OCAgICAgICBkOCAgICAgLGFkUFBQUFA4OCA4YiAgICAgICAgICAKICA4OCwgICI4YSwgICAsYWEgODhiLCAgICxhOCIgODg4IDg4LCAgICAsODggIjhhLCAgICxhYSAgCiAgIlk4ODggYCJZYmJkOCInIDg4YFliYmRQIicgIDg4OCBgIjhiYmRQIlk4ICBgIlliYmQ4IicgIAogICAgICAgICAgICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgODggICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCg=="

func makeDirectories() {
	log.Trace().Msgf("establishing log directory presence at %s...", config.LogDir)
	err := os.MkdirAll(config.LogDir, 0o644)
	if err != nil {
		log.Fatal().
			Str("directory", config.LogDir).
			Err(err).Msg("failed to open log directory")
		return
	}

	log.Trace().Msgf("establishing data directory presence at %s...", config.DBDir)
	err = os.MkdirAll(config.DBDir, 0o644)
	if err != nil {
		log.Fatal().
			Str("directory", config.DBDir).
			Err(err).Msg("failed to open directory")
		return
	}
}

func printBanner() {
	out := termenv.String(b64d(Banner))
	p := termenv.ColorProfile()
	out = out.Foreground(p.Color("#948DB8"))
	fmt.Println(out)
}

func catchSignal() {
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c:
			log.Warn().Msg("Interrupt detected, shutting down gracefully...")
			err := db.SyncAndCloseAll()
			if err != nil {
				log.Warn().Err(err).Msg("sync failure!")
			}
			os.Exit(0)
		}
	}
}

func main() {
	printBanner()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(consoleWriter).With().Timestamp().Logger()
	log.Info().Msg("Initializing...")
	config.Init()
	makeDirectories()
	log.Debug().Msg("debug enabled")
	log.Trace().Msg("trace enabled")
	lf, err := os.OpenFile(config.LogDir+"tcpac.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
	if err != nil {
		log.Fatal().Str("config.LogDir", config.LogDir).Err(err).Msg("Error opening log file!")
	}
	multi := zerolog.MultiLevelWriter(consoleWriter, lf)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	err = dbInit()
	if err != nil {
		log.Fatal().Err(err).Msg("bitcask failure")
	}
	defer db.SyncAndCloseAll()
	go serveTermbin()
	go httpRouter()
	catchSignal()
}
