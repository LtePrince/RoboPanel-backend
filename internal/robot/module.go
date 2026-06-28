package robot

import (
	"go.uber.org/fx"

	"RoboPanel-backend/internal/robot/service"
)

var Module = fx.Module("robot",
	fx.Provide(service.NewRobotService),
)
