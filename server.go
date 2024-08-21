package tapmond

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/jessevdk/go-flags"
	"github.com/lightninglabs/lndclient"
	"github.com/lightninglabs/taproot-assets/taprpc"
	"github.com/lightningnetwork/lnd/build"
	"github.com/lightningnetwork/lnd/lncfg"
	"github.com/lightningnetwork/lnd/macaroons"
	"github.com/lightningnetwork/lnd/signal"
	"github.com/tapmon/tapmond/mons"
	"github.com/tapmon/tapmond/tapmonrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"gopkg.in/macaroon.v2"
)

type TapmondConfig struct {
	lndClient  *lndclient.GrpcLndServices
	tapasConn  *grpc.ClientConn
	tapmonRpc  *TapmonRpcServer
	monManager *mons.Manager
}

type Tapmond struct {
	cfg *TapmondConfig

	rpcServer *TapmonRpcServer

	wg sync.WaitGroup
}

func InitTapmond() (*Tapmond, error) {
	return &Tapmond{
		cfg: &TapmondConfig{},
	}, nil
}

func (t *Tapmond) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
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

	// Create our lnd client.
	lndclient, err := lndclient.NewLndServices(&lndclient.LndServicesConfig{
		LndAddress:         config.Lnd.Host,
		TLSPath:            config.Lnd.TLSPath,
		CustomMacaroonPath: config.Lnd.MacaroonPath,
		Network:            lndclient.Network(config.Network),
	})
	if err != nil {
		return err
	}
	lndGi, err := lndclient.Client.GetInfo(ctx)
	if err != nil {
		return fmt.Errorf("unable to connect to lnd: %v", err)
	}
	defer lndclient.Close()
	log.Infof("Connected to lnd %x", lndGi.IdentityPubkey)
	t.cfg.lndClient = lndclient

	tapasConn, err := getGrpcConnection(
		config.TaprootAssets.Host, config.TaprootAssets.TLSPath, config.TaprootAssets.MacaroonPath,
	)
	if err != nil {
		return err
	}

	t.cfg.tapasConn = tapasConn

	// Test the taproot assets connection.
	tapClient := taprpc.NewTaprootAssetsClient(tapasConn)
	tapGI, err := tapClient.GetInfo(ctx, &taprpc.GetInfoRequest{})
	if err != nil {
		return fmt.Errorf("unable to connect to taproot assets: %v", err)
	}

	log.Infof("Connected to taproot assets client with version %v", tapGI.Version)

	// Create the tapmon manager.
	t.cfg.monManager = mons.NewManager(tapasConn, lndclient.ChainNotifier, lndclient.ChainKit)

	// Create the tapmon grpc server.
	t.rpcServer = NewTapmonRpcServer(t.cfg.monManager)

	// Start the tapmon grpc server.
	grpcServer := grpc.NewServer()
	tapmonrpc.RegisterTapmonServer(grpcServer, t.rpcServer)
	log.Infof("Tapmon server listening on %v", config.RPCListen)
	listener, err := net.Listen("tcp", config.RPCListen)
	if err != nil {
		return err
	}
	defer listener.Close()
	t.wg.Add(1)
	go func() {
		defer t.wg.Done()
		if err := grpcServer.Serve(listener); err != nil {
			log.Errorf("Tapmon server failed to serve: %v", err)
		}
	}()

	// t.wg.Add(1)
	// go func() {

	// }()

	// t.wg.Wait()
	log.Infof("Tapmond running, waiting for cancel")
	<-shutdownInterceptor.ShutdownChannel()
	grpcServer.GracefulStop()
	t.wg.Wait()
	log.Infof("Tapmond shutdown complete")

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

// getGrpcConnection returns a connection to the gRPC server of the taproot assets daemon.
func getGrpcConnection(host, tlspath, macaroonpath string) (*grpc.ClientConn, error) {
	// First get the TLS credentials for the connection.
	tlsCreds, err := credentials.NewClientTLSFromFile(tlspath, "")
	if err != nil {
		return nil, fmt.Errorf("unable to read TLS credentials: %v", err)
	}

	// Load the macaroon file.
	macBytes, err := ioutil.ReadFile(macaroonpath)
	if err != nil {
		return nil, fmt.Errorf("unable to read macaroon file: %v", err)
	}

	var m macaroon.Macaroon
	if err = m.UnmarshalBinary(macBytes); err != nil {
		return nil, fmt.Errorf("unable to unmarshal macaroon: %v", err)
	}
	// Create the macaroon credentials.
	macCreds, err := macaroons.NewMacaroonCredential(&m)
	if err != nil {
		return nil, fmt.Errorf("unable to create macaroon credentials: %v", err)
	}

	// Dial the gRPC server.
	conn, err := grpc.Dial(host, grpc.WithTransportCredentials(tlsCreds),
		grpc.WithPerRPCCredentials(macCreds))
	if err != nil {
		return nil, fmt.Errorf("unable to dial gRPC server: %v", err)
	}

	return conn, nil
}
