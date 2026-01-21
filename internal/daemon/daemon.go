// Package daemon provides cross-platform service management.
// Supports Windows Service, macOS launchd, and Linux systemd.
package daemon

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/kardianos/service"
)

// Config holds service configuration.
type Config struct {
	Name        string
	DisplayName string
	Description string
}

// DefaultConfig returns the default service configuration.
func DefaultConfig() *Config {
	return &Config{
		Name:        "fitwatch",
		DisplayName: "FitWatch",
		Description: "Watches local FIT files and syncs to Intervals.icu",
	}
}

// Program implements the service.Interface.
type Program struct {
	config *Config
	logger *slog.Logger
	run    func(ctx context.Context) error
	cancel context.CancelFunc
}

// NewProgram creates a new service program.
func NewProgram(cfg *Config, logger *slog.Logger, run func(ctx context.Context) error) *Program {
	return &Program{
		config: cfg,
		logger: logger,
		run:    run,
	}
}

// Start is called when the service starts.
func (p *Program) Start(s service.Service) error {
	p.logger.Info("service starting")
	ctx, cancel := context.WithCancel(context.Background())
	p.cancel = cancel

	go func() {
		if err := p.run(ctx); err != nil && err != context.Canceled {
			p.logger.Error("service error", "error", err)
		}
	}()

	return nil
}

// Stop is called when the service stops.
func (p *Program) Stop(s service.Service) error {
	p.logger.Info("service stopping")
	if p.cancel != nil {
		p.cancel()
	}
	return nil
}

// Service wraps the kardianos service with our config.
type Service struct {
	svc     service.Service
	program *Program
	logger  *slog.Logger
}

// New creates a new service wrapper.
func New(cfg *Config, logger *slog.Logger, run func(ctx context.Context) error) (*Service, error) {
	prg := NewProgram(cfg, logger, run)

	svcConfig := &service.Config{
		Name:        cfg.Name,
		DisplayName: cfg.DisplayName,
		Description: cfg.Description,
	}

	svc, err := service.New(prg, svcConfig)
	if err != nil {
		return nil, fmt.Errorf("create service: %w", err)
	}

	return &Service{
		svc:     svc,
		program: prg,
		logger:  logger,
	}, nil
}

// Install installs the service.
func (s *Service) Install() error {
	err := s.svc.Install()
	if err != nil {
		return fmt.Errorf("install service: %w", err)
	}
	s.logger.Info("service installed", "name", s.program.config.Name)
	return nil
}

// Uninstall removes the service.
func (s *Service) Uninstall() error {
	err := s.svc.Uninstall()
	if err != nil {
		return fmt.Errorf("uninstall service: %w", err)
	}
	s.logger.Info("service uninstalled", "name", s.program.config.Name)
	return nil
}

// Start starts the installed service.
func (s *Service) Start() error {
	err := s.svc.Start()
	if err != nil {
		return fmt.Errorf("start service: %w", err)
	}
	s.logger.Info("service started", "name", s.program.config.Name)
	return nil
}

// Stop stops the running service.
func (s *Service) Stop() error {
	err := s.svc.Stop()
	if err != nil {
		return fmt.Errorf("stop service: %w", err)
	}
	s.logger.Info("service stopped", "name", s.program.config.Name)
	return nil
}

// Restart restarts the service.
func (s *Service) Restart() error {
	err := s.svc.Restart()
	if err != nil {
		return fmt.Errorf("restart service: %w", err)
	}
	s.logger.Info("service restarted", "name", s.program.config.Name)
	return nil
}

// Status returns the service status.
func (s *Service) Status() (string, error) {
	status, err := s.svc.Status()
	if err != nil {
		return "", fmt.Errorf("get status: %w", err)
	}

	switch status {
	case service.StatusRunning:
		return "running", nil
	case service.StatusStopped:
		return "stopped", nil
	default:
		return "unknown", nil
	}
}

// Run runs the service (blocking).
// If running interactively (not as a service), runs directly.
// If running as a service, integrates with the service manager.
func (s *Service) Run() error {
	return s.svc.Run()
}

// IsInteractive returns true if running interactively (not as a service).
func (s *Service) IsInteractive() bool {
	return service.Interactive()
}

// Platform returns the current service platform.
func Platform() string {
	return service.Platform()
}

// IsWindowsService returns true if we can install as a Windows service.
func IsWindowsService() bool {
	return service.Platform() == "windows-service"
}

// IsLaunchd returns true if we can install as a macOS launchd service.
func IsLaunchd() bool {
	return service.Platform() == "darwin-launchd"
}

// IsSystemd returns true if we can install as a Linux systemd service.
func IsSystemd() bool {
	return service.Platform() == "linux-systemd"
}

// RunningAsService returns true if currently running as a service (not interactive).
func RunningAsService() bool {
	return !service.Interactive()
}

// PrintInstallHelp prints platform-specific installation help.
func PrintInstallHelp() {
	fmt.Println("Service Management Commands:")
	fmt.Println()
	fmt.Println("  fitwatch service install    Install as a system service")
	fmt.Println("  fitwatch service uninstall  Remove the system service")
	fmt.Println("  fitwatch service start      Start the service")
	fmt.Println("  fitwatch service stop       Stop the service")
	fmt.Println("  fitwatch service restart    Restart the service")
	fmt.Println("  fitwatch service status     Show service status")
	fmt.Println()
	fmt.Printf("Platform: %s\n", Platform())
	fmt.Println()

	switch {
	case IsWindowsService():
		fmt.Println("Windows Notes:")
		fmt.Println("  - Run 'fitwatch service install' as Administrator")
		fmt.Println("  - Service runs as Local System by default")
		fmt.Println("  - View in Services.msc as 'FitWatch'")
	case IsLaunchd():
		fmt.Println("macOS Notes:")
		fmt.Println("  - Service installed to ~/Library/LaunchAgents/")
		fmt.Println("  - Runs as current user")
		fmt.Println("  - Use 'launchctl list | grep fitwatch' to verify")
	case IsSystemd():
		fmt.Println("Linux Notes:")
		fmt.Println("  - Run 'fitwatch service install' as root or with sudo")
		fmt.Println("  - Service unit installed to /etc/systemd/system/")
		fmt.Println("  - Use 'systemctl status fitwatch' to check")
	}
}

// GetLogPath returns the recommended log path for the service.
func GetLogPath() string {
	switch {
	case IsWindowsService():
		// Windows Event Log or app data
		appData := os.Getenv("PROGRAMDATA")
		if appData == "" {
			appData = "C:\\ProgramData"
		}
		return appData + "\\FitWatch\\fitwatch.log"
	case IsLaunchd():
		home, _ := os.UserHomeDir()
		return home + "/Library/Logs/fitwatch.log"
	case IsSystemd():
		// journald handles logging, but provide a file option
		return "/var/log/fitwatch/fitwatch.log"
	default:
		home, _ := os.UserHomeDir()
		return home + "/.fitwatch/fitwatch.log"
	}
}
