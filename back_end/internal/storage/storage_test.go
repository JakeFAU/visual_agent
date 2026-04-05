package storage

import (
	"testing"

	"github.com/JakeFAU/visual_agent/internal/graph"
)

func TestStorageSaveAndLoadGraph(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	input := graph.Graph{
		Version: graph.SupportedGraphVersion,
		Name:    "Test Graph",
		Nodes:   []graph.Node{},
		Edges:   []graph.Edge{},
	}

	if err := store.Save(input); err != nil {
		t.Fatalf("failed to save graph: %v", err)
	}

	loaded, err := store.Load(input.Name)
	if err != nil {
		t.Fatalf("failed to load graph: %v", err)
	}

	if loaded.Name != input.Name {
		t.Fatalf("loaded graph name mismatch: got %q want %q", loaded.Name, input.Name)
	}
}

func TestStorageRejectsPathTraversalGraphNames(t *testing.T) {
	dir := t.TempDir()
	store, err := New(dir)
	if err != nil {
		t.Fatalf("failed to create storage: %v", err)
	}

	input := graph.Graph{
		Version: graph.SupportedGraphVersion,
		Name:    "../outside",
	}

	if err := store.Save(input); err == nil {
		t.Fatal("expected path traversal error, got nil")
	}

	if _, err := store.Load("../outside"); err == nil {
		t.Fatal("expected path traversal load error, got nil")
	}
}
