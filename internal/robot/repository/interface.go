package repository

import "robot-panel/internal/robot/schema"

type IRobotRepository interface {
	GetState() schema.RobotState
}
