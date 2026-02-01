package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/marcelblijleven/tesla-delivery-tui/internal/api"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/config"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/storage"
	"github.com/marcelblijleven/tesla-delivery-tui/internal/tui"
)

// Version information (set by goreleaser via ldflags)
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	// Parse flags
	demoMode := flag.Bool("demo", false, "Run in demo mode with mock data")
	showVersion := flag.Bool("version", false, "Show version information")
	watchMode := flag.Bool("watch", false, "Auto-refresh every 5 minutes")
	watchInterval := flag.Duration("interval", 5*time.Minute, "Auto-refresh interval (e.g., 10m, 1h)")
	flag.Parse()

	if *showVersion {
		fmt.Printf("tesla-delivery-tui %s\n", version)
		fmt.Printf("  commit: %s\n", commit)
		fmt.Printf("  built:  %s\n", date)
		os.Exit(0)
	}

	// Initialize config
	cfg, err := config.New()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing config: %v\n", err)
		os.Exit(1)
	}

	// Initialize API client
	client := api.NewClient(cfg)

	// Initialize history storage
	history, err := storage.NewHistory(cfg.ConfigDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing history storage: %v\n", err)
		os.Exit(1)
	}

	// Initialize checklist storage
	checklist, err := storage.NewChecklist(cfg.ConfigDir())
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error initializing checklist storage: %v\n", err)
		os.Exit(1)
	}

	// Create the TUI model
	model := tui.New(cfg, client, history, checklist)
	if *demoMode {
		model = model.WithDemoMode()
	}
	if *watchMode {
		model = model.WithAutoRefresh(*watchInterval)
	}

	// Run the program with mouse support
	p := tea.NewProgram(model,
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
	)
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error running program: %v\n", err)
		os.Exit(1)
	}
}
