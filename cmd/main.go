package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"RoboPanel-backend/internal/camera"
	"RoboPanel-backend/internal/config"
	"RoboPanel-backend/internal/record"
	"RoboPanel-backend/internal/robot"
	"RoboPanel-backend/internal/ros"
	"RoboPanel-backend/internal/server"
)

var (
	BuildTime = "unknown"
	GitCommit = "unknown"
	GoVersion = "unknown"
)

func main() {
	var cfgFile string

	serveCmd := &cobra.Command{
		Use:   "serve",
		Short: "Start the robot panel API server",
		Run: func(cmd *cobra.Command, args []string) {
			app := fx.New(
				config.Module(cfgFile),
				ros.Module,
				robot.Module,
				record.Module,
				camera.Module,
				server.Module,
				fx.Invoke(func(*server.HttpServer) {}),
			)
			app.Run()
		},
	}

	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Show version info",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("robot-panel\n  build:  %s\n  commit: %s\n  go:     %s\n",
				BuildTime, GitCommit, GoVersion)
		},
	}

	rootCmd := &cobra.Command{
		Use:   "robot-panel",
		Short: "Robot monitoring and data collection API server",
		Run: func(cmd *cobra.Command, args []string) {
			serveCmd.Run(cmd, args)
		},
	}

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "configs/config.yaml", "config file path")
	rootCmd.AddCommand(serveCmd, versionCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
