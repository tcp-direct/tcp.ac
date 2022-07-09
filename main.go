package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog/log"

	"git.tcp.direct/tcp.direct/tcp.ac/config"
)

var Banner string = "CiAgLGQgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogIDg4ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKTU04OE1NTSAsYWRQUFliYSwgOGIsZFBQWWJhLCAgICAgICxhZFBQWVliYSwgICxhZFBQWWJhLCAgCiAgODggICBhOCIgICAgICIiIDg4UCcgICAgIjhhICAgICAiIiAgICAgYFk4IGE4IiAgICAgIiIgIAogIDg4ICAgOGIgICAgICAgICA4OCAgICAgICBkOCAgICAgLGFkUFBQUFA4OCA4YiAgICAgICAgICAKICA4OCwgICI4YSwgICAsYWEgODhiLCAgICxhOCIgODg4IDg4LCAgICAsODggIjhhLCAgICxhYSAgCiAgIlk4ODggYCJZYmJkOCInIDg4YFliYmRQIicgIDg4OCBgIjhiYmRQIlk4ICBgIlliYmQ4IicgIAogICAgICAgICAgICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgODggICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCg=="

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

func printBanner() {
	out := termenv.String(b64d(Banner))
	p := termenv.ColorProfile()
	out = out.Foreground(p.Color("#948DB8"))
	fmt.Println(out)
}

func waitFor(router *gin.Engine) {
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	for {
		select {
		case <-c:
			log.Warn().Msg("Interrupt detected, shutting down gracefully...")
			router.
			return
		}
	}
}

func main() {
	config.Init()
	printBanner()
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
	go serveTermbin()
	waitFor(httpRouter())
}
