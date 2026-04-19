// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package main

import (
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"

	"study/picoclaw/cmd/picoclaw/internal"
	"study/picoclaw/cmd/picoclaw/internal/agent"
	"study/picoclaw/cmd/picoclaw/internal/auth"
	"study/picoclaw/cmd/picoclaw/internal/cron"
	"study/picoclaw/cmd/picoclaw/internal/gateway"
	"study/picoclaw/cmd/picoclaw/internal/model"
	"study/picoclaw/cmd/picoclaw/internal/onboard"
	"study/picoclaw/cmd/picoclaw/internal/skills"
	"study/picoclaw/cmd/picoclaw/internal/status"
	"study/picoclaw/cmd/picoclaw/internal/version"
	"study/picoclaw/pkg/config"
	"study/picoclaw/pkg/updater"
)

func init() {
}

func NewPicoclawCommand() *cobra.Command {
	short := fmt.Sprintf("%s picoclaw - Personal AI Assistant %s\n\n", internal.Logo, config.GetVersion())

	cmd := &cobra.Command{
		Use:     "picoclaw",
		Short:   short,
		Example: "picoclaw version",
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		skills.NewSkillsCommand(),
		model.NewModelCommand(),
		updater.NewUpdateCommand("picoclaw"),
		version.NewVersionCommand(),
	)

	return cmd
}

func main() {
	tz_env := os.Getenv("TZ")
	if tz_env != "" {
		fmt.Println("TZ environment:", tz_env)
		zoneinfo_env := os.Getenv("ZONEINFO")
		fmt.Println("ZONEINFO environment:", zoneinfo_env)
		loc, err := time.LoadLocation(tz_env)
		if err != nil {
			fmt.Println("Error loading time zone:", err)
		} else {
			fmt.Println("Time zone loaded successfully:", loc)
			time.Local = loc //nolint:gosmopolitan // We intentionally set local timezone from TZ env
		}
	}

	cmd := NewPicoclawCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
