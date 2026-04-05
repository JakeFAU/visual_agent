package storage

import (
	"encoding/json"
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"os"
	"path/filepath"
	"strings"
)

type Storage struct {
	dataDir string
}

// New initializes a graph storage rooted at dir, creating the directory if it
// does not already exist.
func New(dir string) (*Storage, error) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(absDir, 0755); err != nil {
		return nil, err
	}
	return &Storage{dataDir: absDir}, nil
}

// Save writes a graph as pretty-printed JSON under its graph name.
func (s *Storage) Save(g graph.Graph) error {
	path, err := s.graphFilePath(g.Name)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(g, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

// Load reads a graph document by name from disk.
func (s *Storage) Load(name string) (*graph.Graph, error) {
	path, err := s.graphFilePath(name)
	if err != nil {
		return nil, err
	}
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

// List returns all stored graph names without their .json extension.
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

// graphFilePath converts a graph name into a safe path rooted inside the
// storage directory.
func (s *Storage) graphFilePath(name string) (string, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return "", fmt.Errorf("graph name cannot be empty")
	}
	if strings.Contains(trimmed, "/") || strings.Contains(trimmed, "\\") {
		return "", fmt.Errorf("graph name cannot contain path separators")
	}

	path := filepath.Join(s.dataDir, trimmed+".json")
	cleanPath := filepath.Clean(path)
	relPath, err := filepath.Rel(s.dataDir, cleanPath)
	if err != nil {
		return "", err
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("graph name resolves outside the storage directory")
	}
	return cleanPath, nil
}
