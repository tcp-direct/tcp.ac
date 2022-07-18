package config

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/viper"
)

const (
	// Title is the name of the application used throughout the configuration process.
	Title = "tcp.ac"
)

var binInfo map[string]string

func init() {
	binInfo = make(map[string]string)

	info, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	for _, v := range info.Settings {
		binInfo[v.Key] = v.Value
	}

	var err error
	home, err = os.UserHomeDir()
	if err != nil {
		println(err.Error())
		os.Exit(1)
	}
	initDefaults()
}

var (
	BaseURL, HTTPPort, HTTPBind, DBDir, LogDir,
	TermbinListen, UnixSocketPath, AdminKey string
	UIDSize, DeleteKeySize, KVMaxKeySizeMB,
	KVMaxValueSizeMB int
	UnixSocketPermissions uint32
	UseUnixSocket         bool
)

var usage = fmt.Sprintf(`
               %s

        brought to you by:
        --> tcp.direct <--

--config <file>		Specify custom config file
--nocolor		Disable color and banner
--genconfig		Write default config to stdout and exit
--version		Show version info and exit
`, Title)

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
		case "--version":
			PrintBanner()
			os.Exit(0)
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
		newconfig := Title
		snek.SetConfigName(newconfig)
		if err = snek.MergeInConfig(); err != nil {
			if err = snek.SafeWriteConfigAs(newconfig + ".toml"); err != nil {
				fmt.Println(err.Error())
				os.Exit(1)
			}
		}
		Filename = newconfig + ".toml"
		return
	}

	if _, err := os.Stat(prefConfigLocation); os.IsNotExist(err) {
		if err = os.MkdirAll(prefConfigLocation, 0o740); err != nil {
			println("error writing new config: " + err.Error())
			os.Exit(1)
		}
	}

	newconfig := prefConfigLocation + "config.toml"
	if err = snek.SafeWriteConfigAs(newconfig); err != nil {
		log.Fatal().Caller().Err(err).Str("target", newconfig).Msg("failed to write new configuration file")
	}

	Filename = newconfig
}

func init() {

}

var once = &sync.Once{}

func substantiateLogger() {
	once.Do(func() {
		consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
		lf, err := os.OpenFile(LogDir+"tcpac.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o666)
		if err != nil {
			log.Fatal().Str("config.LogDir", LogDir).Err(err).Msg("Error opening log file!")
		}
		multi := zerolog.MultiLevelWriter(consoleWriter, lf)
		log.Logger = zerolog.New(multi).With().Timestamp().Logger()
	})
}

// Init will initialize our toml configuration engine and define our default configuration values which can be written to a new configuration file if desired
func Init() {
	argParse()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	consoleWriter := zerolog.ConsoleWriter{Out: os.Stdout}
	log.Logger = log.Output(consoleWriter).With().Timestamp().Logger()
	prefConfigLocation = home + "/.config/" + Title + "/"
	snek = viper.New()

	if genConfig {
		setDefaults()
		os.Exit(0)
	}

	snek.SetConfigType("toml")
	snek.SetConfigName("config")

	if customconfig {
		associateExportedVariables()
		substantiateLogger()
		return
	}

	setDefaults()

	for _, loc := range getConfigPaths() {
		snek.AddConfigPath(loc)
	}

	if err := snek.MergeInConfig(); err != nil {
		substantiateLogger()
		log.Warn().Err(err).Msg("failed to read configuration file")
		writeConfig()
	}

	if len(Filename) < 1 {
		Filename = snek.ConfigFileUsed()
	}

	substantiateLogger()
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

// TODO: use this?
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
		log.Fatal().Err(err1).Msg("config file read fatal error during i/o")
	case err2 != nil:
		log.Fatal().Err(err2).Msg("config file read fatal error during parsing")
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
		"data.directory":        &DBDir,
		"logger.directory":      &LogDir,
		"other.termbin_listen":  &TermbinListen,
		"other.base_url":        &BaseURL,
		"admin.key":             &AdminKey,
	}

	if !strings.HasSuffix(BaseURL, "/") {
		BaseURL += "/"
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
		"data.max_key_size":     &KVMaxKeySizeMB,
		"data.max_value_size":   &KVMaxValueSizeMB,
		"other.uid_size":        &UIDSize,
		"other.delete_key_size": &DeleteKeySize,
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
