package main

import (
	"github.com/rs/zerolog/log"
	"github.com/rs/zerolog"
	"fmt"
	"os"
)

func init() {
	// initialize the logger before the config: that way we can output debug lines
        // pertaining to the parsing of the configuration init

	fmt.Println("Initializing...")

	//////////// init logging ////////////

	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr}) // before we read config, do Stderr pretty print (need logs location)

	log.Info().Msg("Reading configuration file...")

	// see config.go
	configRead()

	// now that we know where to put the log file, we can start output (replace logger)

	err  = os.MkdirAll(logDir, 0755)
	if err != nil {
		log.Fatal().Str("directory",logDir).Str("intent","logDir").Err(err)
		return
	}

	err  = os.MkdirAll(dbDir, 0755)
	if err != nil {
		log.Fatal().Str("directory",dbDir).Str("intent","dbDir").Err(err)
		return
	}

	lf, err := os.OpenFile(logDir+"tcpac.log", os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
	if err != nil {
		fmt.Println("Error opening log file: " + err.Error())
	}

	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	multi := zerolog.MultiLevelWriter(consoleWriter, lf)
	log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	dbInit()
}

func main() {
	// see router.go
	httpRouter()
}
