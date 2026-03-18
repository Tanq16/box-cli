package utils

import (
	"fmt"
	"os"

	"charm.land/lipgloss/v2"
	"github.com/rs/zerolog/log"
)

var (
	infoStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	errorStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	warnStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
)

func PrintInfo(pkg, msg string) {
	if GlobalDebugFlag {
		log.Info().Str("package", pkg).Msg(msg)
	} else if GlobalForAIFlag {
		fmt.Println("[INFO] " + msg)
	} else {
		fmt.Println(infoStyle.Render("→ " + msg))
	}
}

func PrintSuccess(pkg, msg string) {
	if GlobalDebugFlag {
		log.Info().Str("package", pkg).Msg(msg)
	} else if GlobalForAIFlag {
		fmt.Println("[OK] " + msg)
	} else {
		fmt.Println(successStyle.Render("✓ " + msg))
	}
}

func PrintError(pkg, msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Error().Str("package", pkg).Err(err).Msg(msg)
		} else {
			log.Error().Str("package", pkg).Msg(msg)
		}
	} else if GlobalForAIFlag {
		fmt.Println("[ERROR] " + msg)
	} else {
		fmt.Println(errorStyle.Render("✗ " + msg))
	}
}

func PrintFatal(pkg, msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Error().Str("package", pkg).Err(err).Msg(msg)
		} else {
			log.Error().Str("package", pkg).Msg(msg)
		}
	} else if GlobalForAIFlag {
		fmt.Println("[ERROR] " + msg)
	} else {
		fmt.Println(errorStyle.Render("✗ " + msg))
	}
	os.Exit(1)
}

func PrintWarn(pkg, msg string, err error) {
	if GlobalDebugFlag {
		if err != nil {
			log.Warn().Str("package", pkg).Err(err).Msg(msg)
		} else {
			log.Warn().Str("package", pkg).Msg(msg)
		}
	} else if GlobalForAIFlag {
		fmt.Println("[WARN] " + msg)
	} else {
		fmt.Println(warnStyle.Render("! " + msg))
	}
}

func PrintGeneric(msg string) {
	fmt.Println(msg)
}
