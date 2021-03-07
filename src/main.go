package main

import (
	"github.com/rs/zerolog/log"
	"github.com/muesli/termenv"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"fmt"
	"os"
)

var Banner string = "ICAsZCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgODggICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIApNTTg4TU1NICxhZFBQWWJhLCA4YixkUFBZYmEsICAgICAgLGFkUFBZWWJhLCAgLGFkUFBZYmEsICAKICA4OCAgIGE4IiAgICAgIiIgODhQJyAgICAiOGEgICAgICIiICAgICBgWTggYTgiICAgICAiIiAgCiAgODggICA4YiAgICAgICAgIDg4ICAgICAgIGQ4ICAgICAsYWRQUFBQUDg4IDhiICAgICAgICAgIAogIDg4LCAgIjhhLCAgICxhYSA4OGIsICAgLGE4IiA4ODggODgsICAgICw4OCAiOGEsICAgLGFhICAKICAiWTg4OCBgIlliYmQ4IicgODhgWWJiZFAiJyAgODg4IGAiOGJiZFAiWTggIGAiWWJiZDgiJyAgCiAgICAgICAgICAgICAgICAgIDg4ICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAogICAgICAgICAgICAgICAgICA4OCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAK"

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

	err  = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatal().Str("directory",logDir).Str("intent","logDir").Err(err).Msg("failed to open directory")
		return
	}

	err  = os.MkdirAll(dbDir, 0755)
	if err != nil {
		log.Fatal().Str("directory",dbDir).Str("intent","dbDir").Err(err).Msg("failed to open directory")
		return
	}

	lf, err := os.OpenFile(logDir+"tcpac.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		log.Fatal().Str("logDir",logDir).Err(err).Msg("Error opening log file!")
	}

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})
	multi := zerolog.MultiLevelWriter(consoleWriter, lf)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	dbInit()
}

func main() {
	// see router.go
	httpRouter()
}
