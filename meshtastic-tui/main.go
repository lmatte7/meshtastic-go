package main

import (
	"flag"
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/lmatte7/meshtastic-tui/internal/ui"
)

func main() {
	// Parse command line flags
	devicePath := flag.String("device", "", "Device path to connect to (e.g., /dev/tty.usbserial-0001)")
	flag.Parse()

	// Create the model
	model, err := ui.NewModel()
	if err != nil {
		fmt.Printf("Error creating model: %v\n", err)
		os.Exit(1)
	}

	// If device path provided, auto-connect
	if *devicePath != "" {
		fmt.Printf("Auto-connecting to device: %s\n", *devicePath)
		model.SetAutoConnect(*devicePath)
	}

	// Create the Bubble Tea program
	p := tea.NewProgram(
		model,
		tea.WithAltScreen(),       // Use alternate screen buffer
		tea.WithMouseCellMotion(), // Enable mouse support
	)

	// Run the program
	if _, err := p.Run(); err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}
}
