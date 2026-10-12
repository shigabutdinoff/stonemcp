package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/spf13/cobra"

	"github.com/shigabutdinoff/stonemcp/internal/cli"
)

var (
	buildVersion string
	buildNumber  string
	buildDate    string
	buildCommit  string
)

func main() {
	exit(run(os.Args[1:], os.LookupEnv, os.Stdin, os.Stdout, os.Stderr))
}

func exit(code int) {
	os.Exit(code)
}

func versionLine() string {
	switch {
	case buildVersion == "":
		return "N/A"
	case buildNumber != "" && buildNumber != "dev":
		return buildVersion + " (" + buildNumber + ")"
	case buildCommit != "":
		return buildVersion + " (" + buildCommit + ")"
	}
	return buildVersion
}

func run(args []string, env lookup, stdin io.ReadCloser, stdout io.WriteCloser, stderr io.Writer) int {
	signal.Ignore(syscall.SIGPIPE)
	j := newJournal(stderr)
	defer j.shutdown()
	code := 0
	root := cli.New(buildVersion, buildNumber, buildDate, buildCommit, stdout, j)
	settle := bindFlags(root, env)
	root.RunE = func(*cobra.Command, []string) error {
		o, err := settle()
		if err != nil {
			return err
		}
		ctx, stop := cancelOnFirstSignal()
		defer stop()
		code = serve(ctx, o, &mcp.IOTransport{Reader: stdin, Writer: stdout}, j)
		return nil
	}
	root.SetArgs(append([]string{}, args...))
	if err := root.Execute(); err != nil {
		fmt.Fprintln(j, "Ошибка:", err)
		return 2
	}
	return code
}

func cancelOnFirstSignal() (context.Context, context.CancelFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	context.AfterFunc(ctx, stop)
	return ctx, stop
}
