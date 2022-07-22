package config

import (
	"io"
	"os"
	"runtime"

	"git.tcp.direct/kayos/common/entropy"
	"github.com/spf13/afero"
)

var (
	configSections = []string{"logger", "http", "data", "other", "admin"}
	defNoColor     = false
)

var defOpts map[string]map[string]interface{}

func initDefaults() {
	defOpts = map[string]map[string]interface{}{
		"logger": {
			"directory":         home + "/.local/share/tcp.ac/logs",
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
			"trusted_proxies":         []string{"127.0.0.1"},
		},
		"data": {
			"directory":         home + "/.local/share/tcp.ac/data",
			"max_key_size_mb":   10,
			"max_value_size_mb": 20,
		},
		"other": {
			"uid_size":        5,
			"delete_key_size": 12,
			"termbin_listen":  "127.0.0.1:9999",
			"base_url":        "http://localhost:8080/",
		},
		"admin": {
			"key": entropy.RandStrWithUpper(24),
		},
	}
}

func gen(memfs afero.Fs) {
	if err := snek.SafeWriteConfigAs("config.toml"); err != nil {
		print(err.Error())
		os.Exit(1)
	}
	var f afero.File
	var err error
	f, err = memfs.Open("config.toml")
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	newcfg, err := io.ReadAll(f)
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	println(string(newcfg))
}

func setDefaults() {
	memfs := afero.NewMemMapFs()
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		snek.SetDefault("logger.directory", "./logs/")
		defNoColor = true
	}
	for _, def := range configSections {
		snek.SetDefault(def, defOpts[def])
	}
	if genConfig {
		snek.SetFs(memfs)
		gen(memfs)
	}
}
