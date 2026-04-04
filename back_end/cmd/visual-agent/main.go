package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/JakeFAU/visual_agent/internal/compiler"
	"github.com/JakeFAU/visual_agent/internal/graph"
	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "visual-agent",
		Short: "Visual Agent CLI - Compile and run AI agent graphs",
	}

	var compileCmd = &cobra.Command{
		Use:   "compile [file]",
		Short: "Compile a JSON graph file into ADK agents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filePath := args[0]
			data, err := os.ReadFile(filePath)
			if err != nil {
				return fmt.Errorf("failed to read file: %w", err)
			}

			var g graph.Graph
			if err := json.Unmarshal(data, &g); err != nil {
				return fmt.Errorf("failed to unmarshal JSON: %w", err)
			}

			c := compiler.New()
			c.Register("llm_node", &compiler.LLMNodeCompiler{})
			c.Register("toolbox", &compiler.ToolboxNodeCompiler{})
			
			// Note: input/output/logic nodes need registration too, 
			// but we'll focus on the core for this scaffolding task.

			compiled, err := c.Compile(g)
			if err != nil {
				return fmt.Errorf("compilation failed: %w", err)
			}

			fmt.Printf("Successfully compiled graph '%s' with %d nodes.\n", g.Name, len(g.Nodes))
			_ = compiled // In a real run, we would pass this to the engine
			
			return nil
		},
	}

	rootCmd.AddCommand(compileCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
