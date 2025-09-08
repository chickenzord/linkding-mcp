package server

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/chickenzord/linkding-mcp/internal/version"
	"github.com/chickenzord/linkding-mcp/pkg/linkding"
	mcpsdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// MCPServer wraps the MCP SDK server
type MCPServer struct {
	linkdingClient *linkding.Client
	mcpServer      *mcpsdk.Server
}

func (s *MCPServer) RunHTTP(ctx context.Context, bindAddress string) error {
	httpHandler := mcpsdk.NewStreamableHTTPHandler(func(r *http.Request) *mcpsdk.Server {
		return s.mcpServer
	}, nil)

	return http.ListenAndServe(bindAddress, httpHandler)
}

func (s *MCPServer) RunStdio(ctx context.Context) error {
	return s.mcpServer.Run(ctx, &mcpsdk.StdioTransport{})
}

func (s *MCPServer) handleSearchBookmarks(ctx context.Context, req *mcpsdk.CallToolRequest, args SearchBookmarksArgs) (*mcpsdk.CallToolResult, any, error) {
	limit := args.Limit
	if limit == 0 {
		limit = 20
	}

	bookmarks, err := s.linkdingClient.GetBookmarks(ctx, limit, 0, args.Query)
	if err != nil {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: fmt.Sprintf("Failed to search bookmarks: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	result := fmt.Sprintf("Found %d bookmarks:\n\n", len(bookmarks.Results))
	for _, bookmark := range bookmarks.Results {
		result += fmt.Sprintf("• **%s**\n  URL: %s\n", bookmark.Title, bookmark.URL)

		if bookmark.Description != "" {
			result += fmt.Sprintf("  Description: %s\n", bookmark.Description)
		}

		if len(bookmark.TagNames) > 0 {
			result += fmt.Sprintf("  Tags: %v\n", bookmark.TagNames)
		}

		result += "\n"
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{
				Text: result,
			},
		},
	}, nil, nil
}

func (s *MCPServer) handleCreateBookmark(ctx context.Context, req *mcpsdk.CallToolRequest, args CreateBookmarkArgs) (*mcpsdk.CallToolResult, BookmarkResult, error) {
	if args.URL == "" {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: "URL is required",
				},
			},
			IsError: true,
		}, BookmarkResult{}, nil
	}

	createReq := linkding.CreateBookmarkRequest{
		URL:         args.URL,
		Title:       args.Title,
		Description: args.Description,
		TagNames:    args.Tags,
	}

	bookmark, err := s.linkdingClient.CreateBookmark(ctx, createReq)
	if err != nil {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: fmt.Sprintf("Failed to create bookmark: %v", err),
				},
			},
			IsError: true,
		}, BookmarkResult{}, nil
	}

	result := fmt.Sprintf("✅ Bookmark created successfully!\n\n• **%s**\n  URL: %s\n  ID: %d",
		bookmark.Title, bookmark.URL, bookmark.ID)

	if bookmark.Description != "" {
		result += fmt.Sprintf("\n  Description: %s", bookmark.Description)
	}

	if len(bookmark.TagNames) > 0 {
		result += fmt.Sprintf("\n  Tags: %v", bookmark.TagNames)
	}

	bookmarkResult := BookmarkResult{
		ID:          bookmark.ID,
		URL:         bookmark.URL,
		Title:       bookmark.Title,
		Description: bookmark.Description,
		Tags:        bookmark.TagNames,
		Success:     true,
		Message:     "Bookmark created successfully",
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{
				Text: result,
			},
		},
	}, bookmarkResult, nil
}

func (s *MCPServer) handleGetTags(ctx context.Context, req *mcpsdk.CallToolRequest, args GetTagsArgs) (*mcpsdk.CallToolResult, any, error) {
	limit := args.Limit
	if limit == 0 {
		limit = 50
	}

	tags, err := s.linkdingClient.GetTags(ctx, limit, 0)
	if err != nil {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: fmt.Sprintf("Failed to get tags: %v", err),
				},
			},
			IsError: true,
		}, nil, nil
	}

	if len(tags.Results) == 0 {
		return &mcpsdk.CallToolResult{
			Content: []mcpsdk.Content{
				&mcpsdk.TextContent{
					Text: "No tags found",
				},
			},
		}, nil, nil
	}

	result := fmt.Sprintf("Found %d tags:\n\n", len(tags.Results))
	for _, tag := range tags.Results {
		result += fmt.Sprintf("• %s (ID: %s)\n", tag.Name, strconv.Itoa(tag.ID))
	}

	return &mcpsdk.CallToolResult{
		Content: []mcpsdk.Content{
			&mcpsdk.TextContent{
				Text: result,
			},
		},
	}, nil, nil
}

// NewMCP creates a new MCP server using the official MCP Go SDK
func NewMCP(linkdingURL, apiToken string) *MCPServer {
	s := &MCPServer{
		linkdingClient: linkding.NewClient(linkdingURL, apiToken),
	}

	// Create MCP server with implementation info
	versionInfo := version.Get()
	mcpServer := mcpsdk.NewServer(&mcpsdk.Implementation{
		Name:    "linkding-mcp",
		Version: versionInfo.Version,
		Title:   "Linkding MCP Server",
	}, nil)

	// Add search_bookmarks tool
	mcpsdk.AddTool(mcpServer, &mcpsdk.Tool{
		Name:        "search_bookmarks",
		Description: "Search bookmarks in Linkding",
	}, s.handleSearchBookmarks)

	// Add create_bookmark tool
	mcpsdk.AddTool(mcpServer, &mcpsdk.Tool{
		Name:        "create_bookmark",
		Description: "Create a new bookmark in Linkding",
	}, s.handleCreateBookmark)

	// Add get_tags tool
	mcpsdk.AddTool(mcpServer, &mcpsdk.Tool{
		Name:        "get_tags",
		Description: "Get all available tags from Linkding",
	}, s.handleGetTags)

	s.mcpServer = mcpServer

	return s
}
