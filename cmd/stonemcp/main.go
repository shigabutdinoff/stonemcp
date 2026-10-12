package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var (
	version   = "dev"
	commit    string
	buildDate string
)

func main() {
	exit(run(os.Args[1:], os.LookupEnv, os.Stdin, os.Stdout, os.Stderr))
}

func exit(code int) {
	os.Exit(code)
}

func versionString() string {
	if commit == "" {
		return version
	}
	return version + " (" + commit + ")"
}

func run(args []string, env lookup, stdin io.ReadCloser, stdout io.WriteCloser, stderr io.Writer) int {
	signal.Ignore(syscall.SIGPIPE)
	j := newJournal(stderr)
	defer j.shutdown()
	o, err := parseFlags(args, env, j)
	switch {
	case errors.Is(err, flag.ErrHelp):
		return 0
	case err != nil:
		return 2
	case o.showVersion:
		fmt.Fprintln(stdout, versionString())
		return 0
	}
	ctx, stop := cancelOnFirstSignal()
	defer stop()
	return serve(ctx, o, &mcp.IOTransport{Reader: stdin, Writer: stdout}, j)
}

func cancelOnFirstSignal() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	context.AfterFunc(ctx, stop)
	return ctx, stop
}
