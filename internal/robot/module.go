package robot

import (
	"go.uber.org/fx"

	"robot-panel/internal/robot/repository"
	"robot-panel/internal/robot/service"
	"robot-panel/internal/ros"
)

var Module = fx.Module("robot",
	fx.Provide(
		func(ros *ros.Client) repository.IRobotRepository {
			return repository.NewRobotRepository(ros)
		},
		service.NewRobotService,
	),
)
