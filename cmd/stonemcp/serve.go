package main

import (
	"cmp"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/shigabutdinoff/stonemcp/internal/engine"
	"github.com/shigabutdinoff/stonemcp/internal/mcpserver"
)

func serve(ctx context.Context, o options, transport mcp.Transport, stderr io.Writer) int {
	logOut := stderr
	if o.LogFile != "" {
		f, err := os.OpenFile(o.LogFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			fmt.Fprintln(stderr, "Ошибка: файл журнала:", err)
			return 1
		}
		defer f.Close()
		logOut = f
	}
	logger := slog.New(slog.NewTextHandler(logOut, nil))
	if err := play(ctx, o, transport, logger); err != nil {
		logger.Error("stonemcp остановлен", "ошибка", err)
		return 1
	}
	return 0
}

func play(ctx context.Context, o options, transport mcp.Transport, logger *slog.Logger) error {
	v := versionLine()
	logger.Info("stonemcp запущен", "версия", cmp.Or(buildVersion, "N/A"), "номер сборки", cmp.Or(buildNumber, "N/A"),
		"дата сборки", cmp.Or(buildDate, "N/A"), "коммит", cmp.Or(buildCommit, "N/A"), "режим", o.Mode, "адрес", o.Addr)
	ctx, cancel := context.WithCancel(ctx)
	eng := engine.New(engine.Config{Addr: o.Addr, Logger: logger})
	var engineDone sync.WaitGroup
	engineDone.Go(func() {
		if err := eng.Run(ctx); err != nil {
			logger.Error("движок остановлен", "ошибка", err)
		}
	})
	server := mcp.NewServer(&mcp.Implementation{Name: "stonemcp", Version: v}, nil)
	mcpserver.Register(server, eng, mcpserver.Options{Version: v})
	err := server.Run(ctx, transport)
	cancel()
	engineDone.Wait()
	if err == nil || errors.Is(err, context.Canceled) {
		return nil
	}
	return fmt.Errorf("%w: обмен с клиентом MCP", err)
}
