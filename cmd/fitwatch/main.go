// FitWatch - Local FIT file watcher with pluggable consumers.
//
// Usage:
//
//	fitwatch                    # Run interactively (watches for files)
//	fitwatch --once             # Sync existing files and exit
//	fitwatch --config           # Show config path
//	fitwatch --init             # Create default config file
//
// Service commands:
//
//	fitwatch service install    # Install as system service
//	fitwatch service uninstall  # Remove system service
//	fitwatch service start      # Start the service
//	fitwatch service stop       # Stop the service
//	fitwatch service restart    # Restart the service
//	fitwatch service status     # Show service status
package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/johnazariah/fitwatch/internal/config"
	"github.com/johnazariah/fitwatch/internal/consumer"
	"github.com/johnazariah/fitwatch/internal/consumer/intervals"
	"github.com/johnazariah/fitwatch/internal/daemon"
	"github.com/johnazariah/fitwatch/internal/store"
	"github.com/johnazariah/fitwatch/internal/watcher"
)

var (
	version = "dev"
	commit  = "none"
)

func main() {
	// Check for service subcommand first
	if len(os.Args) > 1 && os.Args[1] == "service" {
		handleServiceCommand(os.Args[2:])
		return
	}

	// Flags
	configPath := flag.String("c", config.DefaultConfigPath(), "config file path")
	showConfig := flag.Bool("config", false, "show config path and exit")
	initConfig := flag.Bool("init", false, "create default config and exit")
	once := flag.Bool("once", false, "sync existing files and exit (no watch)")
	verbose := flag.Bool("v", false, "verbose logging")
	showVersion := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	// Setup logging
	logLevel := slog.LevelInfo
	if *verbose {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))
	slog.SetDefault(logger)

	// Handle simple flags
	if *showVersion {
		fmt.Printf("fitwatch %s (%s)\n", version, commit)
		return
	}

	if *showConfig {
		fmt.Println(*configPath)
		return
	}

	if *initConfig {
		if err := initConfigFile(*configPath); err != nil {
			logger.Error("failed to create config", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Created config at: %s\n", *configPath)
		fmt.Println("Edit the file to add your Intervals.icu credentials.")
		return
	}

	// If running as a service, use the service runner
	if daemon.RunningAsService() {
		runAsService(*configPath, logger)
		return
	}

	// Interactive mode
	if *once {
		runOnce(*configPath, logger)
	} else {
		runInteractive(*configPath, logger)
	}
}

func handleServiceCommand(args []string) {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	configPath := config.DefaultConfigPath()

	if len(args) == 0 {
		daemon.PrintInstallHelp()
		return
	}

	// Create the service
	svc, err := daemon.New(
		daemon.DefaultConfig(),
		logger,
		func(ctx context.Context) error {
			return runWatcher(ctx, configPath, logger)
		},
	)
	if err != nil {
		logger.Error("failed to create service", "error", err)
		os.Exit(1)
	}

	switch args[0] {
	case "install":
		if err := svc.Install(); err != nil {
			logger.Error("install failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service installed successfully.")
		fmt.Println("Start with: fitwatch service start")

	case "uninstall":
		if err := svc.Uninstall(); err != nil {
			logger.Error("uninstall failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service uninstalled successfully.")

	case "start":
		if err := svc.Start(); err != nil {
			logger.Error("start failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service started.")

	case "stop":
		if err := svc.Stop(); err != nil {
			logger.Error("stop failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service stopped.")

	case "restart":
		if err := svc.Restart(); err != nil {
			logger.Error("restart failed", "error", err)
			os.Exit(1)
		}
		fmt.Println("Service restarted.")

	case "status":
		status, err := svc.Status()
		if err != nil {
			logger.Error("status failed", "error", err)
			os.Exit(1)
		}
		fmt.Printf("Service status: %s\n", status)

	default:
		fmt.Printf("Unknown service command: %s\n", args[0])
		daemon.PrintInstallHelp()
		os.Exit(1)
	}
}

func runAsService(configPath string, logger *slog.Logger) {
	svc, err := daemon.New(
		daemon.DefaultConfig(),
		logger,
		func(ctx context.Context) error {
			return runWatcher(ctx, configPath, logger)
		},
	)
	if err != nil {
		logger.Error("failed to create service", "error", err)
		os.Exit(1)
	}

	if err := svc.Run(); err != nil {
		logger.Error("service failed", "error", err)
		os.Exit(1)
	}
}

func runInteractive(configPath string, logger *slog.Logger) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		logger.Info("shutting down...")
		cancel()
	}()

	if err := runWatcher(ctx, configPath, logger); err != nil && err != context.Canceled {
		logger.Error("watcher error", "error", err)
		os.Exit(1)
	}
}

func runOnce(configPath string, logger *slog.Logger) {
	ctx := context.Background()

	cfg, syncStore, dispatcher, err := setup(configPath, logger)
	if err != nil {
		logger.Error("setup failed", "error", err)
		os.Exit(1)
	}

	// Handle new FIT files
	handleNewFile := makeFileHandler(ctx, dispatcher, syncStore, logger)

	// Create watcher
	w := watcher.New(cfg.WatchDirs, handleNewFile, logger)

	// Just scan existing files
	logger.Info("scanning for existing FIT files...")
	if err := w.ScanExisting(); err != nil {
		logger.Error("scan failed", "error", err)
		os.Exit(1)
	}
	logger.Info("done")
}

func runWatcher(ctx context.Context, configPath string, logger *slog.Logger) error {
	cfg, syncStore, dispatcher, err := setup(configPath, logger)
	if err != nil {
		return err
	}
	_ = syncStore // TODO: wire up properly

	// Handle new FIT files
	handleNewFile := makeFileHandler(ctx, dispatcher, syncStore, logger)

	// Create watcher
	w := watcher.New(cfg.WatchDirs, handleNewFile, logger)

	// Scan existing files first
	logger.Info("scanning for existing FIT files...")
	if err := w.ScanExisting(); err != nil {
		logger.Warn("scan failed", "error", err)
	}

	// Start watching
	logger.Info("watching for new FIT files", "dirs", cfg.WatchDirs)
	return w.Watch(ctx)
}

func setup(configPath string, logger *slog.Logger) (*config.Config, *store.Store, *consumer.Dispatcher, error) {
	// Load config
	cfg, err := config.LoadOrCreate(configPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("load config: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, nil, nil, fmt.Errorf("invalid config: %w", err)
	}

	// Setup store
	storePath := cfg.StorePath
	if storePath == "" {
		storePath = config.DefaultStorePath()
	}
	syncStore, err := store.New(storePath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("open store: %w", err)
	}

	// Setup consumers
	dispatcher := consumer.NewDispatcher()

	if cfg.Intervals.Enabled {
		ic := intervals.New(cfg.Intervals.AthleteID, cfg.Intervals.APIKey)
		dispatcher.AddConsumer(ic)
		logger.Info("enabled consumer", "name", ic.Name())
	}

	if err := dispatcher.ValidateAll(); err != nil {
		return nil, nil, nil, fmt.Errorf("consumer validation failed: %w", err)
	}

	return cfg, syncStore, dispatcher, nil
}

func makeFileHandler(ctx context.Context, dispatcher *consumer.Dispatcher, syncStore *store.Store, logger *slog.Logger) func(string) {
	return func(path string) {
		results := dispatcher.Dispatch(ctx, path)

		for _, r := range results {
			if r.Success {
				logger.Info("synced", "path", r.FitPath, "consumer", r.Consumer)
				// TODO: Record in store with full metadata
			} else {
				logger.Error("sync failed", "path", r.FitPath, "consumer", r.Consumer, "error", r.Error)
			}
		}
	}
}

func initConfigFile(path string) error {
	cfg := config.DefaultConfig()
	cfg.Intervals.Enabled = false
	cfg.Intervals.AthleteID = "i12345" // Placeholder
	cfg.Intervals.APIKey = "your-api-key-here"
	return cfg.Save(path)
}
