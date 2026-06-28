package service

import (
	"context"

	"robot-panel/internal/robot/repository"
	"robot-panel/internal/robot/schema"
)

type RobotService struct {
	repo repository.IRobotRepository
}

func NewRobotService(repo repository.IRobotRepository) *RobotService {
	return &RobotService{repo: repo}
}

type GetStateReq struct{}

func (s *RobotService) GetState(_ context.Context, _ *GetStateReq) (*schema.RobotState, error) {
	state := s.repo.GetState()
	return &state, nil
}
