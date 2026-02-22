package utils

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/rs/zerolog/log"
)

var (
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // blue
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11")) // yellow
)

// PrintInfo prints an info message in blue
func PrintInfo(pkg, msg string) {
	if GlobalDebugFlag {
		log.Info().Str("package", pkg).Msg(msg)
	} else {
		fmt.Println(infoStyle.Render("→ " + msg))
	}
}

// PrintSuccess prints a success message in green
func PrintSuccess(pkg, msg string) {
	if GlobalDebugFlag {
		log.Info().Str("package", pkg).Msg(msg)
	} else {
		fmt.Println(successStyle.Render("✓ " + msg))
	}
}

// PrintError prints an error message in red (does not exit)
// When debug is enabled, also logs the actual error
func PrintError(pkg, msg string, err error) {
	if GlobalDebugFlag && err != nil {
		log.Error().Str("package", pkg).Err(err).Msg(msg)
	} else {
		fmt.Println(errorStyle.Render("✗ " + msg))
	}
}

// PrintFatal prints an error message and exits
// When debug is enabled, also logs the actual error
func PrintFatal(pkg, msg string, err error) {
	if GlobalDebugFlag && err != nil {
		log.Error().Str("package", pkg).Err(err).Msg(msg)
	} else {
		fmt.Println(errorStyle.Render("✗ " + msg))
	}
	os.Exit(1)
}

// PrintWarn prints a warning message in yellow
// When debug is enabled, also logs the actual error
func PrintWarn(pkg, msg string, err error) {
	if GlobalDebugFlag && err != nil {
		log.Warn().Str("package", pkg).Err(err).Msg(msg)
	} else {
		fmt.Println(warnStyle.Render("! " + msg))
	}
}

// PrintGeneric prints plain text without styling
func PrintGeneric(msg string) {
	fmt.Println(msg)
}
