package runtime

import (
	"context"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/adk/session/vertexai"
	"google.golang.org/genai"
	"iter"
)

// AgentRuntime abstracts the underlying model and session provider
type AgentRuntime interface {
	// Execute runs a compiled agent with the given input string
	Execute(ctx context.Context, a agent.Agent, input string) iter.Seq2[*session.Event, error]
}

// BaseRuntime provides a common implementation using the ADK Runner
type BaseRuntime struct {
	sessionService session.Service
}

// newUserMessage normalizes CLI and HTTP input into the ADK content shape
// expected by runner.Run.
func newUserMessage(input string) *genai.Content {
	return genai.NewContentFromText(input, genai.RoleUser)
}

// Execute runs a compiled agent against the configured session service and
// returns the ADK event stream unchanged.
func (r *BaseRuntime) Execute(ctx context.Context, a agent.Agent, input string) iter.Seq2[*session.Event, error] {
	adkRunner, err := runner.New(runner.Config{
		AppName:           "VisualAgent",
		Agent:             a,
		SessionService:    r.sessionService,
		AutoCreateSession: true,
	})
	if err != nil {
		return func(yield func(*session.Event, error) bool) {
			yield(nil, err)
		}
	}

	return adkRunner.Run(ctx, "default-user", "default-session", newUserMessage(input), agent.RunConfig{})
}

// NewVertexRuntime returns a runtime powered by Vertex AI session storage.
func NewVertexRuntime(ctx context.Context, projectID, location string) (AgentRuntime, error) {
	svc, err := vertexai.NewSessionService(ctx, vertexai.VertexAIServiceConfig{
		ProjectID: projectID,
		Location:  location,
	})
	if err != nil {
		return nil, err
	}
	return &BaseRuntime{sessionService: svc}, nil
}

// NewLocalRuntime returns an in-memory runtime suitable for local development
// and tests.
func NewLocalRuntime() AgentRuntime {
	return &BaseRuntime{sessionService: session.InMemoryService()}
}
