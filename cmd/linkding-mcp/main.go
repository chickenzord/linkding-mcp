package main

import (
	"context"
	"fmt"
	"os"

	"github.com/chickenzord/linkding-mcp/internal/mcp"
)

func main() {
	ctx := context.Background()
	
	linkdingURL := os.Getenv("LINKDING_URL")
	apiToken := os.Getenv("LINKDING_API_TOKEN")
	
	if linkdingURL == "" || apiToken == "" {
		fmt.Fprintf(os.Stderr, "Error: LINKDING_URL and LINKDING_API_TOKEN environment variables are required\n")
		os.Exit(1)
	}
	
	server := mcp.NewServer(linkdingURL, apiToken)
	
	if err := server.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error running MCP server: %v\n", err)
		os.Exit(1)
	}
}