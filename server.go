package tapmond

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/lightningnetwork/lnd/build"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/signal"
)

type Tapmond struct {
	rpcServer *TapmonRpcServer

	wg sync.WaitGroup
}

func InitTapmond() (*Tapmond, error) {
	return &Tapmond{}, nil
}

func (t *Tapmond) Run() error {
	config := DefaultConfig()

	// Parse command line flags.
	parser := flags.NewParser(&config, flags.Default)
	parser.SubcommandsOptional = true

	_, err := parser.Parse()
	if e, ok := err.(*flags.Error); ok && e.Type == flags.ErrHelp {
		return nil
	}
	if err != nil {
		return err
	}

	// Parse ini file.
	tapmonDir := lncfg.CleanAndExpandPath(config.TapmonDir)
	configFile, hasExplicitConfig := getConfigPath(config, tapmonDir)

	if err := flags.IniParse(configFile, &config); err != nil {
		// File not existing is OK as long as it wasn't specified
		// explicitly. All other errors (parsing, EACCESS...) indicate
		// misconfiguration and need to be reported. In case of
		// non-not-found FS errors there's high likelihood that other
		// operations in data directory would also fail so we treat it
		// as early detection of a problem.
		if hasExplicitConfig || !os.IsNotExist(err) {
			return err
		}
	}

	// Parse command line flags again to restore flags overwritten by ini
	// parse.
	_, err = parser.Parse()
	if err != nil {
		return err
	}

	// Show the version and exit if the version flag was specified.
	appName := filepath.Base(os.Args[0])
	appName = strings.TrimSuffix(appName, filepath.Ext(appName))
	if config.ShowVersion {
		fmt.Println(appName, "version", Version())
		os.Exit(0)
	}

	// Special show command to list supported subsystems and exit.
	if config.DebugLevel == "show" {
		fmt.Printf("Supported subsystems: %v\n",
			logWriter.SupportedSubsystems())
		os.Exit(0)
	}

	// Validate our config before we proceed.
	if err := Validate(&config); err != nil {
		return err
	}

	// Start listening for signal interrupts regardless of which command
	// we are running. When our command tries to get a lnd connection, it
	// blocks until lnd is synced. We listen for interrupts so that we can
	// shutdown the daemon while waiting for sync to complete.
	shutdownInterceptor, err := signal.Intercept()
	if err != nil {
		return err
	}

	// Initialize logging at the default logging level.
	logWriter := build.NewRotatingLogWriter()
	SetupLoggers(logWriter, shutdownInterceptor)

	err = logWriter.InitLogRotator(
		filepath.Join(config.LogDir, defaultLogFilename),
		config.MaxLogFileSize, config.MaxLogFiles,
	)
	if err != nil {
		return err
	}
	err = build.ParseAndSetDebugLevels(config.DebugLevel, logWriter)
	if err != nil {
		return err
	}

	// Print the version before executing either primary directive.
	log.Infof("Version: %v", Version())

	// t.wg.Add(1)
	// go func() {

	// }()

	// t.wg.Wait()
	log.Infof("Tapmond running, waiting for cancel")
	<-shutdownInterceptor.ShutdownChannel()
	return nil
}

// getConfigPath gets our config path based on the values that are set in our
// config. The returned bool is set to true if the config file path was set
// explicitly by the user and thus should not be ignored if it doesn't exist.
func getConfigPath(cfg Config, tapmonDir string) (string, bool) {
	// If the config file path provided by the user is set, then we just
	// use this value.
	if cfg.ConfigFile != defaultConfigFile {
		return lncfg.CleanAndExpandPath(cfg.ConfigFile), true
	}

	// If the user has set a tapmon directory that is different to the default
	// we will use this tapmon directory as the location of our config file.
	// We do not namespace by network, because this is a custom tapmon dir.
	if tapmonDir != TapmonDirBase {
		return filepath.Join(tapmonDir, defaultConfigFilename), false
	}

	// Otherwise, we are using our default tapmon directory, and the user did
	// not set a config file path. We use our default tapmon dir, namespaced
	// by network.
	return filepath.Join(tapmonDir, cfg.Network, defaultConfigFilename), false
}
