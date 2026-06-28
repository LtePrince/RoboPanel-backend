package service

import (
	"context"

	"RoboPanel-backend/internal/robot/schema"
	"RoboPanel-backend/internal/ros"
)

type RobotService struct {
	ros *ros.Client
}

func NewRobotService(ros *ros.Client) *RobotService {
	return &RobotService{ros: ros}
}

type GetStateReq struct{}

func (s *RobotService) GetState(_ context.Context, _ *GetStateReq) (*schema.RobotState, error) {
	state := s.ros.GetState()
	return &state, nil
}
