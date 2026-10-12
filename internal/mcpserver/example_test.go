package mcpserver_test

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/mcpserver"
)

func ExampleRegister() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	eng := engine.New(engine.Config{})
	var running sync.WaitGroup
	running.Go(func() {
		if err := eng.Run(ctx); err != nil {
			fmt.Println(err)
		}
	})
	server := mcp.NewServer(&mcp.Implementation{Name: "stonemcp", Version: "dev"}, nil)
	mcpserver.Register(server, eng, mcpserver.Options{Version: "dev"})
	err := server.Run(ctx, &mcp.StdioTransport{})
	stop()
	running.Wait()
	if err != nil {
		fmt.Println(err)
	}
}
