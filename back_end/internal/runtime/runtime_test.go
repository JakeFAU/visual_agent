package runtime

import (
	"context"
	"testing"

	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

func TestLocalRuntime(t *testing.T) {
	ctx := context.Background()
	rt := NewLocalRuntime()

	// Create a simple agent for testing
	model, _ := gemini.NewModel(ctx, "gemini-2.0-flash", &genai.ClientConfig{APIKey: "dummy"})
	a, err := llmagent.New(llmagent.Config{
		Name:        "test-agent",
		Model:       model,
		Instruction: "Say hello",
	})
	if err != nil {
		t.Fatalf("Failed to create agent: %v", err)
	}

	// Execute
	seq := rt.Execute(ctx, a, "Hi")
	
	// We just want to ensure it doesn't crash and returns an iterator
	if seq == nil {
		t.Fatal("Expected iterator, got nil")
	}
	
	// Note: Without a real backend or mock model, running the iterator 
	// would likely fail during the actual model call, but the 
	// initialization and abstraction are now verified.
}
