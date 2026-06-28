package main

import (
	"os"

	"github.com/spf13/cobra"
	"go.uber.org/fx"

	"robot-panel/internal/config"
	"robot-panel/internal/record"
	"robot-panel/internal/robot"
	"robot-panel/internal/ros"
	"robot-panel/internal/server"
)

func main() {
	var cfgFile string

	root := &cobra.Command{
		Use:   "robot-panel",
		Short: "Robot monitoring and data collection API server",
		RunE: func(cmd *cobra.Command, args []string) error {
			app := fx.New(
				config.Module(cfgFile),
				ros.Module,
				robot.Module,
				record.Module,
				server.Module,
				fx.Invoke(func(*server.HttpServer) {}),
			)
			app.Run()
			return nil
		},
	}

	root.Flags().StringVarP(&cfgFile, "config", "c", "configs/config.yaml", "config file path")

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
