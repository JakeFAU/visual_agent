package compiler

import (
	"fmt"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/mcptoolset"
	"os/exec"
)

type ToolboxNodeCompiler struct{}

func (c *ToolboxNodeCompiler) Compile(node graph.Node, metadata map[string]interface{}) (interface{}, error) {
	cfg, ok := node.Config.(graph.ToolboxNodeConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for toolbox")
	}

	var toolsets []tool.Toolset

	// 1. MCP Servers
	for _, mcpCfg := range cfg.MCPServers {
		cmd := exec.Command(mcpCfg.Command, mcpCfg.Args...)
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

	// 2. Custom Functions (Declarations only for now)
	// In v0, we'll focus on MCP and Built-ins.

	return toolsets, nil
}
