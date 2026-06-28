package repository

import (
	"robot-panel/internal/robot/schema"
	"robot-panel/internal/ros"
)

type robotRepository struct {
	ros *ros.Client
}

func NewRobotRepository(ros *ros.Client) IRobotRepository {
	return &robotRepository{ros: ros}
}

func (r *robotRepository) GetState() schema.RobotState {
	return r.ros.GetState()
}
