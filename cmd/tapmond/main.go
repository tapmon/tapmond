package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/tapmon/tapmond"
)

// main starts the lightning-terminal application.
func main() {
	tapmond, err := tapmond.InitTapmond()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	err = tapmond.Run()
	var flagErr *flags.Error
	isFlagErr := errors.As(err, &flagErr)
	if err != nil && (!isFlagErr || flagErr.Type != flags.ErrHelp) {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
