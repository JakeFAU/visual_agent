package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"github.com/JakeFAU/visual_agent/internal/graph"
)

type Storage struct {
	dataDir string
}

func New(dir string) (*Storage, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: dir}, nil
}

func (s *Storage) Save(g graph.Graph) error {
    if g.Name == "" {
        return fmt.Errorf("graph name cannot be empty")
    }
	path := filepath.Join(s.dataDir, g.Name+".json")
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (s *Storage) Load(name string) (*graph.Graph, error) {
	path := filepath.Join(s.dataDir, name+".json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var g graph.Graph
	if err := json.Unmarshal(data, &g); err != nil {
		return nil, err
	}
	return &g, nil
}

func (s *Storage) List() ([]string, error) {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return nil, err
	}
	var names []string
	for _, f := range files {
		if !f.IsDir() && filepath.Ext(f.Name()) == ".json" {
			names = append(names, f.Name()[:len(f.Name())-5])
		}
	}
	return names, nil
}
