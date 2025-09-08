package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chickenzord/linkding-mcp/internal/server"
	"github.com/chickenzord/linkding-mcp/internal/version"
)

func main() {
	ctx := context.Background()

	bindAddr := os.Getenv("BIND_ADDR")
	linkdingURL := os.Getenv("LINKDING_URL")
	apiToken := os.Getenv("LINKDING_API_TOKEN")

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <mode>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Modes: stdio, http, version\n")
		os.Exit(1)
	}

	mode := os.Args[1]

	if mode == "version" {
		versionInfo := version.Get()
		fmt.Println(versionInfo.String())
		os.Exit(0)
	}

	if bindAddr == "" {
		bindAddr = ":8080"
	}

	if linkdingURL == "" || apiToken == "" {
		fmt.Fprintf(os.Stderr, "Error: LINKDING_URL and LINKDING_API_TOKEN environment variables are required\n")
		os.Exit(1)
	}

	mcpServer := server.NewMCP(linkdingURL, apiToken)

	switch mode {
	case "http":
		fmt.Printf("Starting Linkding-MCP HTTP server on %s\n", bindAddr)

		if err := mcpServer.RunHTTP(ctx, bindAddr); err != nil {
			fmt.Fprintf(os.Stderr, "Error running MCP server: %v\n", err)
			os.Exit(1)
		}
	case "stdio":
		if err := mcpServer.RunStdio(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Error running MCP server: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown mode %s\n", mode)
		os.Exit(1)
	}
}
