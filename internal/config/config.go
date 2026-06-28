package config

import (
	"github.com/spf13/viper"
	"go.uber.org/fx"
)

type Config struct {
	Server struct {
		Port int `mapstructure:"port"`
	} `mapstructure:"server"`
	ROS struct {
		MasterURI      string `mapstructure:"master_uri"`
		NodeName       string `mapstructure:"node_name"`
		StaleTimeoutMs int    `mapstructure:"stale_timeout_ms"`
		Topics         struct {
			JointState    string `mapstructure:"joint_state"`
			ToolPose      string `mapstructure:"tool_pose"`
			Odometry      string `mapstructure:"odometry"`
			GalileoStatus string `mapstructure:"galileo_status"`
		} `mapstructure:"topics"`
	} `mapstructure:"ros"`
	Record struct {
		DemoDir    string `mapstructure:"demo_dir"`
		ScriptPath string `mapstructure:"script_path"`
	} `mapstructure:"record"`
	Camera struct {
		Serial      string `mapstructure:"serial"`
		Width       int    `mapstructure:"width"`
		Height      int    `mapstructure:"height"`
		FPS         int    `mapstructure:"fps"`
		JpegQuality int    `mapstructure:"jpeg_quality"`
		TimeoutMs   int    `mapstructure:"timeout_ms"`
	} `mapstructure:"camera"`
}

func load(file string) (*Config, error) {
	viper.SetConfigFile(file)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	var cfg Config
	return &cfg, viper.Unmarshal(&cfg)
}

func Module(file string) fx.Option {
	return fx.Provide(func() (*Config, error) {
		return load(file)
	})
}
