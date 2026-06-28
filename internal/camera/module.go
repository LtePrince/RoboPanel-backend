package camera

import (
	"go.uber.org/fx"

	"RoboPanel-backend/internal/config"
)

var Module = fx.Module("camera",
	fx.Provide(func(cfg *config.Config) Config {
		return Config{
			Serial:      cfg.Camera.Serial,
			Width:       cfg.Camera.Width,
			Height:      cfg.Camera.Height,
			FPS:         cfg.Camera.FPS,
			JpegQuality: cfg.Camera.JpegQuality,
			TimeoutMs:   cfg.Camera.TimeoutMs,
		}
	}),
)
