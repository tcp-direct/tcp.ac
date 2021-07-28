package main

import (
	"fmt"
	"github.com/prologic/bitcask"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"os"
	"strconv"
)

////////////// global declarations
var (
	// datastores
	imgDB  *bitcask.Bitcask
	hashDB *bitcask.Bitcask
	keyDB  *bitcask.Bitcask
	urlDB  *bitcask.Bitcask
	txtDB  *bitcask.Bitcask

	// config directives
	debugBool         bool
	baseUrl           string
	webPort           string
	webIP             string
	dbDir             string
	logDir            string
	uidSize           int
	keySize           int
	txtPort           string

	// utilitarian globals
	err error
	fn  string
	s   string
	i   int
	f   *os.File
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

	baseUrl = viper.GetString("http.baseurl")

	i := viper.GetInt("http.port")
	webPort = strconv.Itoa(i)

	webIP = viper.GetString("http.bindip")
	dbDir = viper.GetString("files.data")
	logDir = viper.GetString("files.logs")
	uidSize = viper.GetInt("global.uidsize")
	keySize = viper.GetInt("global.delkeysize")
	txtPort = viper.GetString("txt.port")

	log.Debug().Str("baseUrl", baseUrl).Str("webIP", webIP).Str("webPort", webPort).Msg("Web")
	log.Debug().Str("logDir", logDir).Str("dbDir", dbDir).Msg("Filesystem")
	log.Debug().Int("keySize", keySize).Int("uidSize", uidSize).Msg("UUIDs")

	//
}
