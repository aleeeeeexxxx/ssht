package main

import (
	"embed"
	"flag"
	"io/fs"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleeeeeexxxx/ssht/internal/api"
	"github.com/aleeeeeexxxx/ssht/internal/config"
	"github.com/aleeeeeexxxx/ssht/internal/logger"
	"github.com/aleeeeeexxxx/ssht/internal/tunnel"
)

//go:embed all:dist
var distFS embed.FS

func main() {
	configPath := flag.String("config", "", "config file path (default: ~/.config/ssht/config.json)")
	debug := flag.Bool("debug", false, "enable debug logging")
	addr := flag.String("addr", ":6001", "http server address")
	flag.Parse()

	logger.Init(*debug)
	defer logger.Sync()

	// Connect logger to WebSocket broadcast
	logger.SetBroadcastFunc(api.BroadcastLog)

	logger.Log.Info("ssht starting")

	cfg, err := config.LoadWithMerge(*configPath, "")
	if err != nil {
		logger.Log.Fatalw("failed to load config", "error", err)
	}

	statePath := config.StatePath()

	logger.Log.Infow("loaded config", "tunnel_count", len(cfg.Tunnels), "state_path", statePath)

	manager := tunnel.NewManager()

	// Add tunnels from config
	for _, tc := range cfg.Tunnels {
		manager.Add(tc)
	}

	// Setup static files
	if subFS, err := fs.Sub(distFS, "dist"); err == nil {
		api.SetStaticFS(subFS)
	}

	// Start HTTP server
	server := api.NewServer(cfg, statePath, manager)

	go func() {
		if err := server.Run(*addr); err != nil {
			logger.Log.Fatalw("http server failed", "error", err)
		}
	}()

	// Wait for interrupt
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh

	logger.Log.Infow("received signal, shutting down", "signal", sig)
	manager.StopAll()
	logger.Log.Info("ssht stopped")
}
