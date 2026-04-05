package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/geminitool"
	"google.golang.org/adk/tool/mcptoolset"
	"os"
	"os/exec"
	"strings"
	"sync"
)

type ToolboxNodeCompiler struct{}

// staticToolset adapts an in-memory slice of tools to the ADK Toolset
// interface.
type staticToolset struct {
	name  string
	tools []tool.Tool
}

func (s *staticToolset) Name() string {
	return s.name
}

// Tools satisfies tool.Toolset for built-ins whose tool list is known at
// compile time.
func (s *staticToolset) Tools(agent.ReadonlyContext) ([]tool.Tool, error) {
	return s.tools, nil
}

// stderrTailBuffer keeps the most recent stderr output from an MCP subprocess
// so lazy initialization failures can surface the actual server-side error.
type stderrTailBuffer struct {
	limit int
	mu    sync.Mutex
	data  []byte
}

func newStderrTailBuffer(limit int) *stderrTailBuffer {
	return &stderrTailBuffer{limit: limit}
}

func (b *stderrTailBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.data = append(b.data, p...)
	if len(b.data) > b.limit {
		b.data = append([]byte(nil), b.data[len(b.data)-b.limit:]...)
	}
	return len(p), nil
}

func (b *stderrTailBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return strings.TrimSpace(string(b.data))
}

// annotatedMCPToolset decorates an MCP-backed toolset with launch details and
// stderr capture so lazy connection failures are actionable.
type annotatedMCPToolset struct {
	inner      tool.Toolset
	serverName string
	command    string
	args       []string
	stderrTail *stderrTailBuffer
}

func (s *annotatedMCPToolset) Name() string {
	return s.inner.Name()
}

func (s *annotatedMCPToolset) Tools(ctx agent.ReadonlyContext) ([]tool.Tool, error) {
	tools, err := s.inner.Tools(ctx)
	if err == nil {
		return tools, nil
	}

	commandLine := s.command
	if len(s.args) > 0 {
		commandLine += " " + strings.Join(s.args, " ")
	}

	stderr := s.stderrTail.String()
	if stderr == "" {
		return nil, fmt.Errorf("%w (mcp server %q command: %s)", err, s.serverName, commandLine)
	}

	return nil, fmt.Errorf("%w\nmcp server %q command: %s\nmcp server stderr:\n%s", err, s.serverName, commandLine, stderr)
}

// Compile resolves built-in tools and external MCP servers referenced by a
// toolbox node into ADK toolsets.
func (c *ToolboxNodeCompiler) Compile(node graph.Node, _ map[string]interface{}) (interface{}, error) {
	cfg, ok := node.Config.(graph.ToolboxNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for toolbox")
	}

	var toolsets []tool.Toolset

	// 1. Built-in tools
	if len(cfg.Tools) > 0 {
		var builtins []tool.Tool
		for _, toolID := range cfg.Tools {
			switch toolID {
			case "google_search":
				builtins = append(builtins, geminitool.GoogleSearch{})
			default:
				return nil, fmt.Errorf("unsupported built-in tool: %s", toolID)
			}
		}

		toolsets = append(toolsets, &staticToolset{
			name:  "builtins",
			tools: builtins,
		})
	}

	// MCP toolsets are command-backed, so each declaration spawns a local
	// subprocess when the toolbox is compiled.
	for _, mcpCfg := range cfg.MCPServers {
		if _, err := exec.LookPath(mcpCfg.Command); err != nil {
			return nil, fmt.Errorf("mcp server %s command %q is not available: %w", mcpCfg.Name, mcpCfg.Command, err)
		}

		cmd := exec.Command(mcpCfg.Command, mcpCfg.Args...)
		cmd.Env = os.Environ()
		stderrTail := newStderrTailBuffer(8 * 1024)
		cmd.Stderr = stderrTail
		// Set environment variables if provided
		if len(mcpCfg.Env) > 0 {
			for k, v := range mcpCfg.Env {
				cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
			}
		}

		mcpTools, err := mcptoolset.New(mcptoolset.Config{
			Transport: &mcp.CommandTransport{Command: cmd},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to initialize MCP server %s: %w", mcpCfg.Name, err)
		}
		toolsets = append(toolsets, &annotatedMCPToolset{
			inner:      mcpTools,
			serverName: mcpCfg.Name,
			command:    mcpCfg.Command,
			args:       append([]string(nil), mcpCfg.Args...),
			stderrTail: stderrTail,
		})
	}

	// 3. Custom Functions (Declarations only for now)
	// In v0, we'll focus on MCP and Built-ins.

	return toolsets, nil
}
