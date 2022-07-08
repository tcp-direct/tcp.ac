package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	// Title is the name of the application used throughout the configuration process.
	Title = "tcp.ac"
)

var (
	// Version roughly represents the applications current version.
	Version = "asdf"
)

var (
	BaseURL, HTTPPort, HTTPBind, DBDir, LogDir,
	TermbinListen, UnixSocketPath string
	UIDSize, DeleteKeySize, KVMaxKeySizeMB,
	KVMaxValueSizeMB int
	UnixSocketPermissions uint32
	UseUnixSocket         bool
)

var usage = []string{
	"\n" + Title + " v" + Version + " Usage\n",
	"-c <toml> - Specify config file",
	"--nocolor - disable color and banner ",
	"--banner - show banner + version and exit",
	"--genconfig - write default config to 'toml' then exit",
}

func printUsage() {
	println(usage)
	os.Exit(0)
}

var (
	forceDebug         = false
	forceTrace         = false
	genConfig          = false
	noColorForce       = false
	customconfig       = false
	home               string
	prefConfigLocation string
	snek               *viper.Viper
)

// TODO: should probably just make a proper CLI with flags or something
func argParse() {
	for i, arg := range os.Args {
		switch arg {
		case "-h":
			printUsage()
		case "--genconfig":
			genConfig = true
		case "--debug", "-v":
			forceDebug = true
		case "--trace", "-vv":
			forceTrace = true
		case "--nocolor":
			noColorForce = true
		case "-c", "--config":
			if len(os.Args) <= i-1 {
				panic("syntax error! expected file after -c")
			}
		default:
			continue
		}
	}
}

// exported generic vars
var (
	// Trace is the value of our trace (extra verbose)  on/off toggle as per the current configuration.
	Trace bool
	// Debug is the value of our debug (verbose) on/off toggle as per the current configuration.
	Debug bool
	// Filename returns the current location of our toml config file.
	Filename string
)

func writeConfig() {
	var err error
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS == "windows" {
		newconfig := "hellpot-config"
		snek.SetConfigName(newconfig)
		if err = snek.MergeInConfig(); err != nil {
			if err = snek.SafeWriteConfigAs(newconfig + ".toml"); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		return
	}

	if _, err := os.Stat(prefConfigLocation); os.IsNotExist(err) {
		if err = os.MkdirAll(prefConfigLocation, 0o750); err != nil {
			println("error writing new config: " + err.Error())
			os.Exit(1)
		}
	}

	newconfig := prefConfigLocation + "/" + "toml"
	if err = snek.SafeWriteConfigAs(newconfig); err != nil {
		fmt.Println("Failed to write new configuration file: " + err.Error())
		os.Exit(1)
	}

	Filename = newconfig
}

// Init will initialize our toml configuration engine and define our default configuration values which can be written to a new configuration file if desired
func Init() {
	argParse()
	prefConfigLocation = home + "/.config/" + Title
	snek = viper.New()

	if genConfig {
		setDefaults()
		println("config file generated at: " + Filename)
		os.Exit(0)
	}

	snek.SetConfigType("toml")
	snek.SetConfigName("config")

	if customconfig {
		associateExportedVariables()
		return
	}

	setDefaults()

	for _, loc := range getConfigPaths() {
		snek.AddConfigPath(loc)
	}

	if err := snek.MergeInConfig(); err != nil {
		println("Error reading configuration file: " + err.Error())
		println("Writing new configuration file...")
		writeConfig()
	}

	if len(Filename) < 1 {
		Filename = snek.ConfigFileUsed()
	}

	associateExportedVariables()
}

func getConfigPaths() (paths []string) {
	paths = append(paths, "./")
	//goland:noinspection GoBoolExpressions
	if runtime.GOOS != "windows" {
		paths = append(paths,
			prefConfigLocation, "/etc/"+Title+"/", "../", "../../")
	}
	return
}

func loadCustomConfig(path string) {
	/* #nosec */
	f, err := os.Open(path)
	if err != nil {
		println("Error opening specified config file: " + path)
		println(err.Error())
		os.Exit(1)
	}

	Filename, _ = filepath.Abs(path)

	if len(Filename) < 1 {
		Filename = path
	}

	defer func(f *os.File) {
		fcerr := f.Close()
		if fcerr != nil {
			fmt.Println("failed to close file handler for config file: ", fcerr.Error())
		}
	}(f)

	buf, err1 := io.ReadAll(f)
	err2 := snek.ReadConfig(bytes.NewBuffer(buf))

	switch {
	case err1 != nil:
		fmt.Println("config file read fatal error during i/o: ", err1.Error())
		os.Exit(1)
	case err2 != nil:
		fmt.Println("config file read fatal error during parse: ", err2.Error())
		os.Exit(1)
	default:
		break
	}

	customconfig = true
}

func processOpts() {
	// string options and their exported variables
	stringOpt := map[string]*string{
		"http.bind_addr":        &HTTPBind,
		"http.bind_port":        &HTTPPort,
		"http.unix_socket_path": &UnixSocketPath,
		"logger.directory":      &LogDir,
		"other.termbin_listen":  &TermbinListen,
	}

	// bool options and their exported variables
	boolOpt := map[string]*bool{
		"http.use_unix_socket": &UseUnixSocket,
		"logger.debug":         &Debug,
		"logger.trace":         &Trace,
		"logger.nocolor":       &noColorForce,
	}

	// integer options and their exported variables
	intOpt := map[string]*int{
		"data.max_key_size":   &KVMaxKeySizeMB,
		"data.max_value_size": &KVMaxValueSizeMB,
		"other.uid_size":      &UIDSize,
	}

	uint32Opt := map[string]*uint32{
		"http.unix_socket_permissions": &UnixSocketPermissions,
	}

	for key, opt := range stringOpt {
		*opt = snek.GetString(key)
	}
	for key, opt := range boolOpt {
		*opt = snek.GetBool(key)
	}
	for key, opt := range intOpt {
		*opt = snek.GetInt(key)
	}
	for key, opt := range uint32Opt {
		*opt = snek.GetUint32(key)
	}
}

func associateExportedVariables() {
	processOpts()
	// We set exported variables here so that it tracks when accessed from other packages.
	if Debug || forceDebug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		Debug = true
	}
	if Trace || forceTrace {
		zerolog.SetGlobalLevel(zerolog.TraceLevel)
		Trace = true
	}
}
