package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"RoboPanel-backend/internal/config"
	"RoboPanel-backend/internal/record/repository"
	"RoboPanel-backend/internal/record/schema"
)

type RecordService struct {
	cfg     *config.Config
	repo    repository.IRecordRepository
	mu      sync.Mutex
	cmd     *exec.Cmd
	logFile *os.File
	pid     int
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
	LogFile string `json:"log_file"`
}

func (s *RecordService) Start(_ context.Context, req *StartReq) (*StartResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return nil, fmt.Errorf("recording already in progress")
	}

	demoDir := fmt.Sprintf("%s/demonstration_%d", s.cfg.Record.DemoDir, req.DemoNum)
	if err := os.MkdirAll(demoDir, 0755); err != nil {
		return nil, fmt.Errorf("create demo dir: %w", err)
	}

	logPath := filepath.Join(demoDir, "record.log")
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, fmt.Errorf("create log file: %w", err)
	}

	cmd := exec.Command("bash", s.cfg.Record.ScriptPath,
		fmt.Sprintf("demo_num=%d", req.DemoNum),
	)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	// Own process group so we can signal the whole tree (all Python subprocesses)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	if err := cmd.Start(); err != nil {
		logFile.Close()
		return nil, fmt.Errorf("start recorder: %w", err)
	}

	s.cmd = cmd
	s.logFile = logFile
	s.pid = cmd.Process.Pid
	s.running = true

	go func() {
		_ = cmd.Wait()
		logFile.Close()
		s.mu.Lock()
		s.running = false
		s.cmd = nil
		s.logFile = nil
		s.mu.Unlock()
	}()

	return &StartResp{
		DemoNum: req.DemoNum,
		DemoDir: demoDir,
		PID:     cmd.Process.Pid,
		LogFile: logPath,
	}, nil
}

// --- Stop ---

type StopReq struct{}

type StopResp struct {
	Message string `json:"message"`
}

func (s *RecordService) Stop(_ context.Context, _ *StopReq) (*StopResp, error) {
	s.mu.Lock()
	if !s.running {
		s.mu.Unlock()
		return nil, fmt.Errorf("no recording in progress")
	}
	pid := s.pid
	s.mu.Unlock()

	// SIGINT to the entire process group: all Python subprocesses (camera,
	// robot state recorders) receive it and flush their files before exiting.
	if err := syscall.Kill(-pid, syscall.SIGINT); err != nil {
		s.mu.Lock()
		if s.cmd != nil {
			_ = s.cmd.Process.Signal(os.Interrupt)
		}
		s.mu.Unlock()
	}

	// Force-kill after 15 s if the process group still hasn't exited.
	go func() {
		time.Sleep(15 * time.Second)
		s.mu.Lock()
		still := s.running
		s.mu.Unlock()
		if still {
			_ = syscall.Kill(-pid, syscall.SIGKILL)
		}
	}()

	return &StopResp{Message: "stopping - data is being saved"}, nil
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
	if s.running {
		resp.PID = s.pid
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

// --- Delete ---

type DeleteReq struct {
	Name string `uri:"name" binding:"required"`
}

type DeleteResp struct {
	Message string `json:"message"`
}

func (s *RecordService) Delete(_ context.Context, req *DeleteReq) (*DeleteResp, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.repo.DeleteDemo(req.Name); err != nil {
		return nil, err
	}
	return &DeleteResp{Message: fmt.Sprintf("deleted %s", req.Name)}, nil
}
