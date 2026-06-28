package repository

import (
	"os"
	"path/filepath"

	"robot-panel/internal/record/schema"
)

type recordRepository struct {
	demoDir string
}

func NewRecordRepository(demoDir string) IRecordRepository {
	return &recordRepository{demoDir: demoDir}
}

func (r *recordRepository) ListDemos() ([]schema.Demo, error) {
	entries, err := os.ReadDir(r.demoDir)
	if os.IsNotExist(err) {
		return []schema.Demo{}, nil
	}
	if err != nil {
		return nil, err
	}

	demos := []schema.Demo{}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		info, _ := e.Info()
		demos = append(demos, schema.Demo{
			Name:      e.Name(),
			CreatedAt: info.ModTime().UnixMilli(),
			Files:     r.listFiles(e.Name()),
		})
	}
	return demos, nil
}

func (r *recordRepository) FileExists(demoName, fileName string) (string, bool) {
	path := filepath.Join(r.demoDir, demoName, fileName)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return "", false
	}
	return path, true
}

func (r *recordRepository) listFiles(demoName string) []schema.DemoFile {
	files := []schema.DemoFile{}
	entries, err := os.ReadDir(filepath.Join(r.demoDir, demoName))
	if err != nil {
		return files
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		files = append(files, schema.DemoFile{Name: e.Name(), Size: info.Size()})
	}
	return files
}
