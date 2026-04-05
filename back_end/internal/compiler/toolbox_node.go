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
		cmd := exec.Command(mcpCfg.Command, mcpCfg.Args...)
		cmd.Env = os.Environ()
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
		toolsets = append(toolsets, mcpTools)
	}

	// 3. Custom Functions (Declarations only for now)
	// In v0, we'll focus on MCP and Built-ins.

	return toolsets, nil
}
