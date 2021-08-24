package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/muesli/termenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"os"
	"os/signal"
	"syscall"
)

var Banner string = "CiAgLGQgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogIDg4ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKTU04OE1NTSAsYWRQUFliYSwgOGIsZFBQWWJhLCAgICAgICxhZFBQWVliYSwgICxhZFBQWWJhLCAgCiAgODggICBhOCIgICAgICIiIDg4UCcgICAgIjhhICAgICAiIiAgICAgYFk4IGE4IiAgICAgIiIgIAogIDg4ICAgOGIgICAgICAgICA4OCAgICAgICBkOCAgICAgLGFkUFBQUFA4OCA4YiAgICAgICAgICAKICA4OCwgICI4YSwgICAsYWEgODhiLCAgICxhOCIgODg4IDg4LCAgICAsODggIjhhLCAgICxhYSAgCiAgIlk4ODggYCJZYmJkOCInIDg4YFliYmRQIicgIDg4OCBgIjhiYmRQIlk4ICBgIlliYmQ4IicgIAogICAgICAgICAgICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAgICAgODggICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCg=="

func init() {
	out := termenv.String(b64d(Banner))
	p := termenv.ColorProfile()
	out = out.Foreground(p.Color("#948DB8"))

	fmt.Println(out)

	// initialize the logger before the config: that way we can output debug lines
	// pertaining to the parsing of the configuration init

	//////////// init logging ////////////

	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	log.Info().Msg("Initializing...")

	// see config.go
	configRead()

	if !debugBool {
		gin.SetMode(gin.ReleaseMode)
	}

	// now that we know where to put the log file, we can start output (replace logger)

	err = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatal().Str("directory", logDir).Str("intent", "logDir").Err(err).Msg("failed to open directory")
		return
	}

	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		log.Fatal().Str("directory", dbDir).Str("intent", "dbDir").Err(err).Msg("failed to open directory")
		return
	}

	lf, err := os.OpenFile(logDir+"tcpac.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Str("logDir", logDir).Err(err).Msg("Error opening log file!")
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	multi := zerolog.MultiLevelWriter(consoleWriter, lf)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	dbInit()
}

func main() {
	defer dbSync()
	go func() {
		c := make(chan os.Signal, 5)
		signal.Notify(c, os.Interrupt, syscall.SIGTERM)
		for {
			select {
			case <-c:
				log.Warn().Msg("Interrupt detected, shutting down gracefully...")
				dbSync()
				os.Exit(0)
			}
		}
	}()
	go serveTermbin()
	// see router.go
	httpRouter()
}
