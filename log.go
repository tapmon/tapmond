package tapmond

import (
	"github.com/btcsuite/btclog"
	"github.com/lightningnetwork/lnd/build"
	"github.com/lightningnetwork/lnd/signal"
)

var (
	// log is a logger that is initialized with no output filters. This
	// means the package will not perform any logging by default until the
	// caller requests it.
	log btclog.Logger

	// interceptor is the OS signal interceptor that we need to keep a
	// reference to for listening for shutdown signals.
	interceptor signal.Interceptor

	logWriter *build.RotatingLogWriter
)

const (
	// Subsystem defines the logging code for this subsystem.
	Subsystem = "TAPMOND"
)

// The default amount of logging is none.
func init() {
	UseLogger(build.NewSubLogger(Subsystem, nil))
}

// UseLogger uses a specified Logger to output package logging info.  This
// should be used in preference to SetLogWriter if the caller is also using
// btclog.
func UseLogger(logger btclog.Logger) {
	log = logger
}

// SetupLoggers initializes all package-global logger variables.
func SetupLoggers(root *build.RotatingLogWriter, intercept signal.Interceptor) {

	logWriter = root
	genLogger := genSubLogger(root, intercept)

	log = build.NewSubLogger(Subsystem, genLogger)
	interceptor = intercept

	// Add the lightning-terminal root logger.
}

// genSubLogger creates a logger for a subsystem. We provide an instance of
// a signal.Interceptor to be able to shutdown in the case of a critical error.
func genSubLogger(root *build.RotatingLogWriter,
	interceptor signal.Interceptor) func(string) btclog.Logger {

	// Create a shutdown function which will request shutdown from our
	// interceptor if it is listening.
	shutdown := func() {
		if !interceptor.Listening() {
			return
		}

		interceptor.RequestShutdown()
	}

	// Return a function which will create a sublogger from our root
	// logger without shutdown fn.
	return func(tag string) btclog.Logger {
		return root.GenSubLogger(tag, shutdown)
	}
}
