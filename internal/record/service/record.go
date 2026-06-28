package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"sync"

	"RoboPanel-backend/internal/config"
	"RoboPanel-backend/internal/record/repository"
	"RoboPanel-backend/internal/record/schema"
)

type RecordService struct {
	cfg     *config.Config
	repo    repository.IRecordRepository
	mu      sync.Mutex
	cmd     *exec.Cmd
	running bool
}

func NewRecordService(cfg *config.Config, repo repository.IRecordRepository) *RecordService {
	return &RecordService{cfg: cfg, repo: repo}
}

// --- Start ---

type StartReq struct {
	DemoNum int `json:"demo_num" binding:"required"`
}

type StartResp struct {
	DemoNum int    `json:"demo_num"`
	DemoDir string `json:"demo_dir"`
	PID     int    `json:"pid"`
}

func (s *RecordService) Start(_ context.Context, req *StartReq) (*StartResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil, fmt.Errorf("recording already in progress")
	}

	demoDir := fmt.Sprintf("%s/demo_%d", s.cfg.Record.DemoDir, req.DemoNum)
	if err := os.MkdirAll(demoDir, 0755); err != nil {
		return nil, fmt.Errorf("create demo dir: %w", err)
	}

	cmd := exec.Command("python3", s.cfg.Record.ScriptPath,
		fmt.Sprintf("demo_num=%d", req.DemoNum),
	)
	cmd.Dir = s.cfg.Record.ScriptWorkdir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start recorder: %w", err)
	}

	s.cmd = cmd
	s.running = true
	go func() {
		_ = cmd.Wait()
		s.mu.Lock()
		s.running = false
		s.cmd = nil
		s.mu.Unlock()
	}()

	return &StartResp{DemoNum: req.DemoNum, DemoDir: demoDir, PID: cmd.Process.Pid}, nil
}

// --- Stop ---

type StopReq struct{}

type StopResp struct {
	Message string `json:"message"`
}

func (s *RecordService) Stop(_ context.Context, _ *StopReq) (*StopResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running || s.cmd == nil {
		return nil, fmt.Errorf("no recording in progress")
	}
	if err := s.cmd.Process.Signal(os.Interrupt); err != nil {
		_ = s.cmd.Process.Kill()
	}
	s.running = false
	s.cmd = nil
	return &StopResp{Message: "recording stopped"}, nil
}

// --- Status ---

type StatusReq struct{}

type StatusResp struct {
	Running bool `json:"running"`
	PID     int  `json:"pid,omitempty"`
}

func (s *RecordService) Status(_ context.Context, _ *StatusReq) (*StatusResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	resp := &StatusResp{Running: s.running}
	if s.running && s.cmd != nil && s.cmd.Process != nil {
		resp.PID = s.cmd.Process.Pid
	}
	return resp, nil
}

// --- List ---

type ListReq struct{}

type ListResp struct {
	Demos []schema.Demo `json:"demos"`
	Total int           `json:"total"`
}

func (s *RecordService) List(_ context.Context, _ *ListReq) (*ListResp, error) {
	demos, err := s.repo.ListDemos()
	if err != nil {
		return nil, err
	}
	return &ListResp{Demos: demos, Total: len(demos)}, nil
}

// FileExists delegates path resolution to the repository.
func (s *RecordService) FileExists(demoName, fileName string) (string, bool) {
	return s.repo.FileExists(demoName, fileName)
}
