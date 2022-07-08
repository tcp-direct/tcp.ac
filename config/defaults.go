package config

import (
	"fmt"
	"os"
	"runtime"
)

var (
	configSections = []string{"logger", "http", "data", "other"}
	defNoColor     = false
)

var defOpts = map[string]map[string]interface{}{
	"logger": {
		"directory":         ".logs",
		"debug":             true,
		"trace":             false,
		"nocolor":           defNoColor,
		"use_date_filename": true,
	},
	"http": {
		"use_unix_socket":         false,
		"unix_socket_path":        "/var/run/tcp.ac",
		"unix_socket_permissions": uint32(0644),
		"bind_addr":               "127.0.0.1",
		"bind_port":               "8080",
	},
	"data": {
		"directory":         ".data",
		"max_key_size_mb":   10,
		"max_value_size_mb": 20,
	},
	"other": {
		"uid_size":        5,
		"delete_key_size": 12,
		"termbin_listen":  "127.0.0.1:9999",
	},
}

func setDefaults() {
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		snek.SetDefault("logger.directory", "./logs/")
		defNoColor = true
	}

	for _, def := range configSections {
		snek.SetDefault(def, defOpts[def])
	}

	if genConfig {
		if err := snek.SafeWriteConfigAs("./config.toml"); err != nil {
			fmt.Println(err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

}
