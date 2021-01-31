package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"strconv"
)

/////////////////////////////////
func configRead() {
	viper.SetConfigName("config") // filename without ext
	viper.SetConfigType("toml")   //  also defines extension

	viper.AddConfigPath("/etc/tcpac/") // multiple possible
	viper.AddConfigPath(".")           //  locations for config

	err = viper.ReadInConfig()
	if err != nil { // this should be replaced with more intelligent handling
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	//// fetch config directives from file ////
	debugBool = viper.GetBool("global.debug") // we need to load the debug boolean first
	//  so we can output config directives
	if debugBool {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		log.Debug().Msg("Debug mode enabled")
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	s = "http.baseurl"
	baseUrl = viper.GetString(s)
	log.Debug().Str(s, baseUrl).Msg("configRead()")

	s = "http.port"
	i := viper.GetInt(s)
	webPort = strconv.Itoa(i)                   // int looks cleaner in config
	log.Debug().Str(s, webPort).Msg("configRead()") //  but we reference it as a string later

	s = "http.bindip"
	webIP = viper.GetString(s)
	log.Debug().Str(s, webIP).Msg("configRead()")

	s = "files.data"
	dbDir = viper.GetString(s)
	log.Debug().Str(s, dbDir).Msg("configRead()") //  where we're actually gonna store everything

	s = "files.logs"
	logDir = viper.GetString(s)
	log.Debug().Str(s, logDir).Msg("configRead()")

	s = "img.uidsize"
	uidSize = viper.GetInt(s)
	log.Debug().Int(s, uidSize).Msg("configRead()")

	s = "img.delkeysize"
	keySize = viper.GetInt(s)
	log.Debug().Int(s, keySize).Msg("configRead()")
}
