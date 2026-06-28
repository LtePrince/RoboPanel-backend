package service

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"robot-panel/internal/config"
	"robot-panel/internal/record/schema"
)

type RecordService struct {
	cfg     *config.Config
	mu      sync.Mutex
	cmd     *exec.Cmd
	running bool
}

func NewRecordService(cfg *config.Config) *RecordService {
	return &RecordService{cfg: cfg}
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

	demoDir := filepath.Join(s.cfg.Record.DemoDir, fmt.Sprintf("demo_%d", req.DemoNum))
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

	return &StartResp{
		DemoNum: req.DemoNum,
		DemoDir: demoDir,
		PID:     cmd.Process.Pid,
	}, nil
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

	// SIGINT first so Python can flush/close HDF5 cleanly
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

// --- List demos ---

type ListReq struct{}

type ListResp struct {
	Demos []schema.Demo `json:"demos"`
	Total int           `json:"total"`
}

func (s *RecordService) List(_ context.Context, _ *ListReq) (*ListResp, error) {
	entries, err := os.ReadDir(s.cfg.Record.DemoDir)
	if os.IsNotExist(err) {
		return &ListResp{Demos: []schema.Demo{}, Total: 0}, nil
	}
	if err != nil {
		return nil, err
	}

	var demos []schema.Demo
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, _ := e.Info()
		demo := schema.Demo{
			Name:      e.Name(),
			CreatedAt: info.ModTime().UnixMilli(),
			Files:     listFiles(filepath.Join(s.cfg.Record.DemoDir, e.Name())),
		}
		demos = append(demos, demo)
	}
	if demos == nil {
		demos = []schema.Demo{}
	}
	return &ListResp{Demos: demos, Total: len(demos)}, nil
}

func listFiles(dir string) []schema.DemoFile {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var files []schema.DemoFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		files = append(files, schema.DemoFile{Name: e.Name(), Size: info.Size()})
	}
	return files
}
