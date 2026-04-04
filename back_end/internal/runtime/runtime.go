package runtime

import (
	"context"
	"google.golang.org/adk/session"
	"iter"
)

// AgentRuntime abstracts the underlying LLM provider (e.g. Vertex AI)
type AgentRuntime interface {
	// Execute runs a compiled agent with the given input
	Execute(ctx context.Context, agent interface{}, input string) iter.Seq2[*session.Event, error]
}
