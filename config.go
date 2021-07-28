package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strconv"
)

func configRead() {
	// name of the file
	// and the extension
	viper.SetConfigName("config")
	viper.SetConfigType("toml")

	// potential config paths
	viper.AddConfigPath("/etc/tcpac/")
	viper.AddConfigPath("../")
	viper.AddConfigPath("./")

	// this should be replaced with more intelligent handling
	err = viper.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	// read config
	debugBool = viper.GetBool("global.debug") // we need to load the debug boolean first
	if debugBool {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug mode enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	// base URL of site
	s = "http.baseurl"
	baseUrl = viper.GetString(s)

	// bind port
	s = "http.port"
	i := viper.GetInt(s)
	webPort = strconv.Itoa(i) // int looks cleaner in config

	// bind IP
	s = "http.bindip"
	webIP = viper.GetString(s)

	// database location (main storage)
	s = "files.data"
	dbDir = viper.GetString(s)

	// logfile location
	s = "files.logs"
	logDir = viper.GetString(s)

	// character count of unique IDs for posts
	s = "img.uidsize"
	uidSize = viper.GetInt(s)

	// size of generated unique delete keys
	s = "img.delkeysize"
	keySize = viper.GetInt(s)

	log.Debug().Str("baseUrl", baseUrl).Str("webIP", webIP).Str("webPort", webPort).Msg("Web")
	log.Debug().Str("logDir", logDir).Str("dbDir", dbDir).Msg("Filesystem")
	log.Debug().Int("keySize", keySize).Int("uidSize", uidSize).Msg("UUIDs")

	//
}
