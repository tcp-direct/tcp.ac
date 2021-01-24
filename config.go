package main

import (
	"github.com/prologic/bitcask"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
	"github.com/rs/zerolog"
	"strconv"
	"fmt"
)

////////////// global declarations
// datastores
var imgDB      *bitcask.Bitcask
var hashDB     *bitcask.Bitcask
var keyDB      *bitcask.Bitcask
var urlDB      *bitcask.Bitcask
var txtDB      *bitcask.Bitcask
// config directives
var debugBool  bool
var baseUrl    string
var webPort    string
var webIP      string
var dbDir      string
var logDir     string
// utilitarian globals
var s	       string
var f	       string
var i	       int
var err        error

////////////////////////////////


func configRead() {
	viper.SetConfigName("config")	     // filename without ext
	viper.SetConfigType("toml")	    //  also defines extension

	viper.AddConfigPath("/etc/tcpac/") // multiple possible
	viper.AddConfigPath(".")	  //  locations for config

	err = viper.ReadInConfig()
	if err != nil {			// this should be replaced with more intelligent handling
		panic(fmt.Errorf("Fatal error reading config file: %s \n", err))
	}

	//// fetch config directives from file ////
	debugBool = viper.GetBool("global.debug")     // we need to load the debug boolean first
						     //  so we can output config directives
	if debugBool {
		log.Debug().Msg("Debug mode enabled")
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	s = "http.baseurl"
	baseUrl = viper.GetString(s)
	log.Debug().Str(s, baseUrl).Msg("[config]")

	s = "http.port"
	i := viper.GetInt(s)
	webPort = strconv.Itoa(i)	      	      // int looks cleaner in config
	log.Debug().Str(s,webPort).Msg("[config]")   //  but we reference it as a string later

	s = "http.bindip"
	webIP = viper.GetString(s)
	log.Debug().Str(s,webIP).Msg("[config]")

	s = "files.data"
	dbDir = viper.GetString(s)
	log.Debug().Str(s,dbDir).Msg("[config]")    //  where we're actually gonna store everything

	s = "files.logs"
	logDir = viper.GetString(s)
	log.Debug().Str(s,logDir).Msg("[config]")
}
