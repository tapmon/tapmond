package tapmond

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/lightningnetwork/lnd/cert"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc/credentials"
)

var (
	// TapmonDirBase is the default main directory where tapmond stores its data.
	TapmonDirBase = btcutil.AppDataDir("tapmond", false)

	// DefaultNetwork is the default bitcoin network tapmond runs on.
	DefaultNetwork = "mainnet"

	defaultLogLevel    = "info"
	defaultLogDirname  = "logs"
	defaultLogFilename = "tapmond.log"

	defaultConfigFilename = "tapmond.conf"

	defaultSqliteDatabaseFileName = "tapmon_sqlite.db"

	defaultLogDir     = filepath.Join(TapmonDirBase, defaultLogDirname)
	defaultConfigFile = filepath.Join(
		TapmonDirBase, DefaultNetwork, defaultConfigFilename,
	)

	// defaultSqliteDatabasePath is the default path under which we store
	// the SQLite database file.
	defaultSqliteDatabasePath = filepath.Join(
		TapmonDirBase, DefaultNetwork, defaultSqliteDatabaseFileName,
	)

	defaultMaxLogFiles    = 3
	defaultMaxLogFileSize = 10

	// DefaultTLSCertFilename is the default file name for the autogenerated
	// TLS certificate.
	DefaultTLSCertFilename = "tls.cert"

	// DefaultTLSKeyFilename is the default file name for the autogenerated
	// TLS key.
	DefaultTLSKeyFilename = "tls.key"

	defaultSelfSignedOrganization = "tapmon autogenerated cert"

	// defaultLndMacaroon is the default macaroon file we use if the old,
	// deprecated --lnd.macaroondir config option is used.
	defaultLndMacaroon = "admin.macaroon"

	// DefaultLndMacaroonPath is the default mainnet admin macaroon path of
	// LND.
	DefaultLndMacaroonPath = filepath.Join(
		btcutil.AppDataDir("lnd", false),
		"data", "chain", "bitcoin", DefaultNetwork,
		defaultLndMacaroon,
	)

	// DefaultLndRPCTimeout is the default timeout to use when communicating
	// with lnd.
	DefaultLndRPCTimeout = time.Minute

	// DefaultTLSCertPath is the default full path of the autogenerated TLS
	// certificate.
	DefaultTLSCertPath = filepath.Join(
		TapmonDirBase, DefaultNetwork, DefaultTLSCertFilename,
	)

	// DefaultTLSKeyPath is the default full path of the autogenerated TLS
	// key.
	DefaultTLSKeyPath = filepath.Join(
		TapmonDirBase, DefaultNetwork, DefaultTLSKeyFilename,
	)

	// DefaultMacaroonFilename is the default file name for the
	// autogenerated tapmon macaroon.
	DefaultMacaroonFilename = "tapmon.macaroon"

	// DefaultMacaroonPath is the default full path of the base tapmon
	// macaroon.
	DefaultMacaroonPath = filepath.Join(
		TapmonDirBase, DefaultNetwork, DefaultMacaroonFilename,
	)

	// DefaultAutogenValidity is the default validity of a self-signed
	// certificate in number of days.
	DefaultAutogenValidity = 365 * 24 * time.Hour
)

type rpcConfig struct {
	Host string `long:"host" description:"lnd instance rpc address"`

	// MacaroonPath is the path to the single macaroon that should be used
	// instead of needing to specify the macaroon directory that contains
	// all of lnd's macaroons. The specified macaroon MUST have all
	// permissions that all the subservers use, otherwise permission errors
	// will occur.
	MacaroonPath string `long:"macaroonpath" description:"The full path to the single macaroon to use, either the admin.macaroon or a custom baked one. Cannot be specified at the same time as macaroondir. A custom macaroon must contain ALL permissions required for all subservers to work, otherwise permission errors will occur."`

	TLSPath string `long:"tlspath" description:"Path to lnd tls certificate"`

	// RPCTimeout is the timeout to use when communicating with lnd.
	RPCTimeout time.Duration `long:"rpctimeout" description:"The timeout to use when communicating with lnd"`
}

type Config struct {
	ShowVersion bool   `long:"version" description:"Display version information and exit"`
	Network     string `long:"network" description:"network to run on" choice:"regtest" choice:"testnet" choice:"mainnet" choice:"simnet"`
	RPCListen   string `long:"rpclisten" description:"Address to listen on for gRPC clients"`
	RESTListen  string `long:"restlisten" description:"Address to listen on for REST clients"`
	CORSOrigin  string `long:"corsorigin" description:"The value to send in the Access-Control-Allow-Origin header. Header will be omitted if empty."`

	TapmonDir  string `long:"tapmondir" description:"The directory for all of tapmon's data. If set, this option overwrites --datadir, --logdir, --tlscertpath, --tlskeypath and --macaroonpath."`
	ConfigFile string `long:"configfile" description:"Path to configuration file."`
	DataDir    string `long:"datadir" description:"Directory for tapmondb."`

	TLSCertPath        string        `long:"tlscertpath" description:"Path to write the TLS certificate for tapmon's RPC and REST services."`
	TLSKeyPath         string        `long:"tlskeypath" description:"Path to write the TLS private key for tapmon's RPC and REST services."`
	TLSExtraIPs        []string      `long:"tlsextraip" description:"Adds an extra IP to the generated certificate."`
	TLSExtraDomains    []string      `long:"tlsextradomain" description:"Adds an extra domain to the generated certificate."`
	TLSAutoRefresh     bool          `long:"tlsautorefresh" description:"Re-generate TLS certificate and key if the IPs or domains are changed."`
	TLSDisableAutofill bool          `long:"tlsdisableautofill" description:"Do not include the interface IPs or the system hostname in TLS certificate, use first --tlsextradomain as Common Name instead, if set."`
	TLSValidity        time.Duration `long:"tlsvalidity" description:"Tapmon's TLS certificate validity period in days. Defaults to 8760h (1 year)"`

	MacaroonPath string `long:"macaroonpath" description:"Path to write the macaroon for tapmon's RPC and REST services if it doesn't exist."`

	LogDir         string `long:"logdir" description:"Directory to log output."`
	MaxLogFiles    int    `long:"maxlogfiles" description:"Maximum logfiles to keep (0 for no rotation)."`
	MaxLogFileSize int    `long:"maxlogfilesize" description:"Maximum logfile size in MB."`

	DebugLevel string `long:"debuglevel" description:"Logging level for all subsystems {trace, debug, info, warn, error, critical} -- You may also specify <subsystem>=<level>,<subsystem2>=<level>,... to set the log level for individual subsystems -- Use show to list available subsystems"`

	Lnd           *rpcConfig `group:"lnd" namespace:"lnd"`
	TaprootAssets *rpcConfig `group:"taprootassets" namespace:"taprootassets"`
}

// DefaultConfig returns all default values for the Config struct.
func DefaultConfig() Config {
	return Config{
		Network:        DefaultNetwork,
		RPCListen:      "localhost:11010",
		RESTListen:     "localhost:8081",
		TapmonDir:      TapmonDirBase,
		ConfigFile:     defaultConfigFile,
		DataDir:        TapmonDirBase,
		LogDir:         defaultLogDir,
		MaxLogFiles:    defaultMaxLogFiles,
		MaxLogFileSize: defaultMaxLogFileSize,
		DebugLevel:     defaultLogLevel,
		TLSCertPath:    DefaultTLSCertPath,
		TLSKeyPath:     DefaultTLSKeyPath,
		TLSValidity:    DefaultAutogenValidity,
		MacaroonPath:   DefaultMacaroonPath,
		Lnd: &rpcConfig{
			Host:         "localhost:10009",
			MacaroonPath: DefaultLndMacaroonPath,
			RPCTimeout:   DefaultLndRPCTimeout,
		},
	}
}

// Validate cleans up paths in the config provided and validates it.
func Validate(cfg *Config) error {
	// Cleanup any paths before we use them.
	cfg.TapmonDir = lncfg.CleanAndExpandPath(cfg.TapmonDir)
	cfg.DataDir = lncfg.CleanAndExpandPath(cfg.DataDir)
	cfg.LogDir = lncfg.CleanAndExpandPath(cfg.LogDir)
	cfg.TLSCertPath = lncfg.CleanAndExpandPath(cfg.TLSCertPath)
	cfg.TLSKeyPath = lncfg.CleanAndExpandPath(cfg.TLSKeyPath)
	cfg.MacaroonPath = lncfg.CleanAndExpandPath(cfg.MacaroonPath)

	// Since our tapmon directory overrides our log/data dir values, make sure
	// that they are not set when tapmon dir is set. We hard here rather than
	// overwriting and potentially confusing the user.
	tapmonDirSet := cfg.TapmonDir != TapmonDirBase

	if tapmonDirSet {
		logDirSet := cfg.LogDir != defaultLogDir
		dataDirSet := cfg.DataDir != TapmonDirBase
		tlsCertPathSet := cfg.TLSCertPath != DefaultTLSCertPath
		tlsKeyPathSet := cfg.TLSKeyPath != DefaultTLSKeyPath

		if logDirSet {
			return fmt.Errorf("tapmondir overwrites logdir, please " +
				"only set one value")
		}

		if dataDirSet {
			return fmt.Errorf("tapmondir overwrites datadir, please " +
				"only set one value")
		}

		if tlsCertPathSet {
			return fmt.Errorf("tapmondir overwrites tlscertpath, " +
				"please only set one value")
		}

		if tlsKeyPathSet {
			return fmt.Errorf("tapmondir overwrites tlskeypath, " +
				"please only set one value")
		}

		// Once we are satisfied that no other config value was set, we
		// replace them with our tapmon dir.
		cfg.DataDir = cfg.TapmonDir
		cfg.LogDir = filepath.Join(cfg.TapmonDir, defaultLogDirname)
	}

	// Append the network type to the data and log directory so they are
	// "namespaced" per network.
	cfg.DataDir = filepath.Join(cfg.DataDir, cfg.Network)
	cfg.LogDir = filepath.Join(cfg.LogDir, cfg.Network)

	// We want the TLS and macaroon files to also be in the "namespaced" sub
	// directory. Replace the default values with actual values in case the
	// user specified either tapmondir or datadir.
	if cfg.TLSCertPath == DefaultTLSCertPath {
		cfg.TLSCertPath = filepath.Join(
			cfg.DataDir, DefaultTLSCertFilename,
		)
	}
	if cfg.TLSKeyPath == DefaultTLSKeyPath {
		cfg.TLSKeyPath = filepath.Join(
			cfg.DataDir, DefaultTLSKeyFilename,
		)
	}
	if cfg.MacaroonPath == DefaultMacaroonPath {
		cfg.MacaroonPath = filepath.Join(
			cfg.DataDir, DefaultMacaroonFilename,
		)
	}

	// If the user doesn't specify Lnd.MacaroonPath, we'll reassemble it
	// with the passed Network options.
	if cfg.Lnd.MacaroonPath == DefaultLndMacaroonPath {
		cfg.Lnd.MacaroonPath = filepath.Join(
			btcutil.AppDataDir("lnd", false),
			"data", "chain", "bitcoin", cfg.Network,
			defaultLndMacaroon,
		)
	}

	// We'll also update the database file location as well, if it wasn't
	// set.
	// if cfg.Sqlite.DatabaseFileName == defaultSqliteDatabasePath {
	// 	cfg.Sqlite.DatabaseFileName = filepath.Join(
	// 		cfg.DataDir, defaultSqliteDatabaseFileName,
	// 	)
	// }

	// If either of these directories do not exist, create them.
	if err := os.MkdirAll(cfg.DataDir, os.ModePerm); err != nil {
		return err
	}

	if err := os.MkdirAll(cfg.LogDir, os.ModePerm); err != nil {
		return err
	}

	// Make sure only one of the macaroon options is used.
	switch {
	case cfg.Lnd.MacaroonPath != "":
		cfg.Lnd.MacaroonPath = lncfg.CleanAndExpandPath(
			cfg.Lnd.MacaroonPath,
		)

	default:
		return fmt.Errorf("must specify --lnd.macaroonpath")
	}

	// TLS Validity period to be at least 24 hours
	if cfg.TLSValidity < time.Hour*24 {
		return fmt.Errorf("TLS certificate minimum validity period is 24h")
	}

	return nil
}

// getTLSConfig generates a new self signed certificate or refreshes an existing
// one if necessary, then returns the full TLS configuration for initializing
// a secure server interface.
func getTLSConfig(cfg *Config) (*tls.Config, *credentials.TransportCredentials,
	error) {

	// Let's load our certificate first or create then load if it doesn't
	// yet exist.
	certData, parsedCert, err := loadCertWithCreate(cfg)
	if err != nil {
		return nil, nil, err
	}

	// If the certificate expired or it was outdated, delete it and the TLS
	// key and generate a new pair.
	if time.Now().After(parsedCert.NotAfter) {
		log.Info("TLS certificate is expired or outdated, " +
			"removing old file then generating a new one")

		err := os.Remove(cfg.TLSCertPath)
		if err != nil {
			return nil, nil, err
		}

		err = os.Remove(cfg.TLSKeyPath)
		if err != nil {
			return nil, nil, err
		}

		certData, _, err = loadCertWithCreate(cfg)
		if err != nil {
			return nil, nil, err
		}
	}

	tlsCfg := cert.TLSConfFromCert(certData)
	tlsCfg.NextProtos = []string{"h2"}
	restCreds, err := credentials.NewClientTLSFromFile(
		cfg.TLSCertPath, "",
	)
	if err != nil {
		return nil, nil, err
	}

	return tlsCfg, &restCreds, nil
}

// loadCertWithCreate tries to load the TLS certificate from disk. If the
// specified cert and key files don't exist, the certificate/key pair is created
// first.
func loadCertWithCreate(cfg *Config) (tls.Certificate, *x509.Certificate,
	error) {

	// Ensure we create TLS key and certificate if they don't exist.
	if !lnrpc.FileExists(cfg.TLSCertPath) &&
		!lnrpc.FileExists(cfg.TLSKeyPath) {

		log.Infof("Generating TLS certificates...")
		certBytes, keyBytes, err := cert.GenCertPair(
			defaultSelfSignedOrganization, cfg.TLSExtraIPs,
			cfg.TLSExtraDomains, cfg.TLSDisableAutofill,
			cfg.TLSValidity,
		)
		if err != nil {
			return tls.Certificate{}, nil, err
		}

		err = cert.WriteCertPair(
			cfg.TLSCertPath, cfg.TLSKeyPath, certBytes, keyBytes,
		)
		if err != nil {
			return tls.Certificate{}, nil, err
		}

		log.Infof("Done generating TLS certificates")
	}

	return cert.LoadCert(cfg.TLSCertPath, cfg.TLSKeyPath)
}
