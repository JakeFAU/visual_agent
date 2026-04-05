package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/JakeFAU/visual_agent/internal/compiler"
	"github.com/JakeFAU/visual_agent/internal/config"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/JakeFAU/visual_agent/internal/runtime"
	"github.com/JakeFAU/visual_agent/internal/server"
	"github.com/JakeFAU/visual_agent/internal/storage"
	"github.com/spf13/cobra"
)

func main() {
	ctx := context.Background()
	cfg, err := config.Load()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	var rootCmd = &cobra.Command{
		Use:   "visual-agent",
		Short: "Visual Agent CLI - Compile and run AI agent graphs",
	}

	var runCmd = &cobra.Command{
		Use:   "run [file] [input]",
		Short: "Compile and execute a JSON graph file",
		Args:  cobra.ExactArgs(2),
		RunE: func(_ *cobra.Command, args []string) error {
			filePath := args[0]
			input := args[1]

			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var g graph.Graph
			if err := json.Unmarshal(data, &g); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}
			if err := g.Validate(); err != nil {
				return fmt.Errorf("invalid graph: %w", err)
			}

			c := compiler.New()
			c.Register("llm_node", &compiler.LLMNodeCompiler{})
			c.Register("toolbox", &compiler.ToolboxNodeCompiler{})
			c.Register("if_else_node", &compiler.IfElseNodeCompiler{})
			c.Register("while_node", &compiler.WhileNodeCompiler{})

			compiledAgent, err := c.Compile(g)
			if err != nil {
				return fmt.Errorf("compilation failed: %w", err)
			}

			var rt runtime.AgentRuntime
			if cfg.RuntimeType == "vertex" {
				rt, err = runtime.NewVertexRuntime(ctx, cfg.ProjectID, cfg.Location)
				if err != nil {
					return fmt.Errorf("failed to initialize vertex runtime: %w", err)
				}
			} else {
				rt = runtime.NewLocalRuntime()
			}

			fmt.Printf("Running graph '%s'...\n", g.Name)
			for event, err := range rt.Execute(ctx, compiledAgent, input) {
				if err != nil {
					return fmt.Errorf("execution error: %w", err)
				}
				if event != nil {
					fmt.Printf("[%s] %v\n", event.Author, event.Content)
				}
			}

			return nil
		},
	}

	var serveCmd = &cobra.Command{
		Use:   "serve",
		Short: "Start the Visual Agent API server",
		RunE: func(_ *cobra.Command, _ []string) error {
			s, err := storage.New("data/graphs")
			if err != nil {
				return err
			}
			srv := server.New(s)
			fmt.Printf("Starting server on %s...\n", cfg.ServerAddr)
			return srv.Run(cfg.ServerAddr)
		},
	}

	rootCmd.AddCommand(runCmd)
	rootCmd.AddCommand(serveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
