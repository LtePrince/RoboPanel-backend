package robot

import (
	"go.uber.org/fx"

	"robot-panel/internal/robot/service"
)

var Module = fx.Module("robot",
	fx.Provide(service.NewRobotService),
)
