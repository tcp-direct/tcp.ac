package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

// //////////// global declarations
var (
	// config directives
	debugBool bool
	baseURL   string
	webPort   string
	webIP     string
	dbDir     string
	logDir    string
	uidSize   int
	keySize   int
	txtPort   string
	maxSize   int

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

	baseURL = viper.GetString("http.baseurl")

	i := viper.GetInt("http.port")
	webPort = strconv.Itoa(i)

	webIP = viper.GetString("http.bindip")
	dbDir = viper.GetString("files.data")
	logDir = viper.GetString("files.logs")
	uidSize = viper.GetInt("global.uidsize")
	keySize = viper.GetInt("global.delkeysize")
	txtPort = viper.GetString("txt.port")
	maxSize = viper.GetInt("files.maxuploadsize")
	//
}
