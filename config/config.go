package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

const (
	// Title is the name of the application used throughout the configuration process.
	Title = "tcp.ac"
)

var (
	// Version roughly represents the applications current version.
	Version = runtime.
)

var (
	BaseURL, WebPort, WebIP, DBDir, LogDir, TxtPort string
	UIDSize, KeySize, MaxSize                       int
)

var usage = []string{
	"\n" + Title + " v" + Version + " Usage\n",
	"-c <config.toml> - Specify config file",
	"--nocolor - disable color and banner ",
	"--banner - show banner + version and exit",
	"--genconfig - write default config to 'config.toml' then exit",
}

func printUsage() {
	println(usage)
	os.Exit(0)
}

var (
	forceDebug = false
	forceTrace = false
)

// TODO: should probably just make a proper CLI with flags or something
func argParse() {
	for i, arg := range os.Args {
		switch arg {
		case "-h":
			printUsage()
		case "--genconfig":
			GenConfig = true
		case "--debug", "-v":
			forceDebug = true
		case "--trace", "-vv":
			forceTrace = true
		case "--nocolor":
			noColorForce = true
		case "--banner":
			BannerOnly = true
		case "-c", "--config":
			if len(os.Args) <= i-1 {
				panic("syntax error! expected file after -c")
			}
			loadCustomConfig(os.Args[i+1])
		default:
			continue
		}
	}
}

// generic vars
var (
	noColorForce       = false
	customconfig       = false
	home               string
	prefConfigLocation string
	snek               *viper.Viper
)

// exported generic vars
var (
	// Trace is the value of our trace (extra verbose)  on/off toggle as per the current configuration.
	Trace bool
	// Debug is the value of our debug (verbose) on/off toggle as per the current configuration.
	Debug bool
	// Filename returns the current location of our toml config file.
	Filename string
)

func init() {
	prefConfigLocation = home + "/.config/" + Title
	snek = viper.New()
}

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

	newconfig := prefConfigLocation + "/" + "config.toml"
	if err = snek.SafeWriteConfigAs(newconfig); err != nil {
		fmt.Println("Failed to write new configuration file: " + err.Error())
		os.Exit(1)
	}

	Filename = newconfig
}

// Init will initialize our toml configuration engine and define our default configuration values which can be written to a new configuration file if desired
func Init() {
	snek.SetConfigType("toml")
	snek.SetConfigName("config")

	argParse()

	if customconfig {
		associateExportedVariables()
		return
	}

	setDefaults()

	for _, loc := range getConfigPaths() {
		snek.AddConfigPath(loc)
	}

	if err = snek.MergeInConfig(); err != nil {
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
	if f, err = os.Open(path); err != nil {
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
		"logger.directory":      &LogDir,
		"deception.server_name": &FakeServerName,
	}
	// string slice options and their exported variables
	strSliceOpt := map[string]*[]string{
		"http.router.paths": &Paths,
	}
	// bool options and their exported variables
	boolOpt := map[string]*bool{
		"performance.restrict_concurrency": &RestrictConcurrency,
		"http.use_unix_socket":             &UseUnixSocket,
		"logger.debug":                     &Debug,
		"logger.trace":                     &Trace,
		"logger.nocolor":                   &NoColor,
		"http.router.makerobots":           &MakeRobots,
		"http.router.catchall":             &CatchAll,
	}
	// integer options and their exported variables
	intOpt := map[string]*int{
		"performance.max_workers": &MaxWorkers,
	}

	for key, opt := range stringOpt {
		*opt = snek.GetString(key)
	}
	for key, opt := range strSliceOpt {
		*opt = snek.GetStringSlice(key)
	}
	for key, opt := range boolOpt {
		*opt = snek.GetBool(key)
	}
	for key, opt := range intOpt {
		*opt = snek.GetInt(key)
	}
}

func associateExportedVariables() {
	processOpts()

	if noColorForce {
		NoColor = true
	}

	if UseUnixSocket {
		UnixSocketPath = snek.GetString("http.unix_socket_path")
		parsedPermissions, err := strconv.ParseUint(snek.GetString("http.unix_socket_permissions"), 8, 32)
		if err == nil {
			UnixSocketPermissions = uint32(parsedPermissions)
		}
	}

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
