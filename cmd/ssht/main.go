package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/aleeeeeexxxx/ssht/internal/api"
	"github.com/aleeeeeexxxx/ssht/internal/config"
	"github.com/aleeeeeexxxx/ssht/internal/logger"
	"github.com/aleeeeeexxxx/ssht/internal/tunnel"
)

func main() {
	configPath := flag.String("config", "", "config file path (default: ~/.config/ssht/config.json)")
	debug := flag.Bool("debug", false, "enable debug logging")
	addr := flag.String("addr", ":6001", "http server address")
	flag.Parse()

	logger.Init(*debug)
	defer logger.Sync()

	logger.Log.Info("ssht starting")

	cfg, err := config.Load(*configPath)
	if err != nil {
		logger.Log.Fatalw("failed to load config", "error", err)
	}

	path := *configPath
	if path == "" {
		path = config.DefaultPath()
	}

	logger.Log.Infow("loaded config", "tunnel_count", len(cfg.Tunnels))

	manager := tunnel.NewManager()

	// Add tunnels from config
	for _, tc := range cfg.Tunnels {
		manager.Add(tc)
	}

	// Start HTTP server
	server := api.NewServer(cfg, path, manager)

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
