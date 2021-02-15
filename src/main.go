package main

import (
	"github.com/rs/zerolog/log"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"encoding/base64"
	"fmt"
	"os"
)

var Banner string = "ICB8fCAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgCiAgfHwgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgIAonJ3x8JycgIC58JycsICd8fCcnfCwgICAgICcnJ3wuICAufCcnLCAKICB8fCAgICB8fCAgICAgfHwgIHx8ICAgIC58Jyd8fCAgfHwgICAgCiAgYHwuLicgYHwuLicgIHx8Li58JyAuLiBgfC4ufHwuIGB8Li4nIAogICAgICAgICAgICAgICB8fCAgICAgICAgICAgICAgICAgICAgICAKICAgICAgICAgICAgICAufHwgICAgICAgICAgICAgICAgICAgICAgCg=="

func b64d(str string) string {
        data, err := base64.StdEncoding.DecodeString(str)
        if err != nil {
                return err.Error()
        }
        return string(data)
}

func init() {
	fmt.Println()
	fmt.Println(b64d(Banner))

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
