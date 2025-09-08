package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/chickenzord/linkding-mcp/pkg/linkding"
)

type Server struct {
	linkdingClient *linkding.Client
}

type MCPRequest struct {
	Jsonrpc string `json:"jsonrpc"`
	ID      any    `json:"id"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type MCPResponse struct {
	Jsonrpc string    `json:"jsonrpc"`
	ID      any       `json:"id"`
	Result  any       `json:"result,omitempty"`
	Error   *MCPError `json:"error,omitempty"`
}

type MCPError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type InitializeParams struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo     `json:"clientInfo"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

func NewServer(linkdingURL, apiToken string) *Server {
	return &Server{
		linkdingClient: linkding.NewClient(linkdingURL, apiToken),
	}
}

func (s *Server) Run(ctx context.Context) error {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	for {
		var request MCPRequest
		if err := decoder.Decode(&request); err != nil {
			return fmt.Errorf("failed to decode request: %w", err)
		}

		response := s.handleRequest(ctx, request)

		if err := encoder.Encode(response); err != nil {
			return fmt.Errorf("failed to encode response: %w", err)
		}
	}
}

func (s *Server) handleRequest(ctx context.Context, request MCPRequest) MCPResponse {
	switch request.Method {
	case "initialize":
		return s.handleInitialize(request)
	case "notifications/initialized":
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
		}
	case "tools/list":
		return s.handleToolsList(request)
	case "tools/call":
		return s.handleToolsCall(ctx, request)
	default:
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Method not found",
			},
		}
	}
}

func (s *Server) handleInitialize(request MCPRequest) MCPResponse {
	return MCPResponse{
		Jsonrpc: "2.0",
		ID:      request.ID,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			ServerInfo: ServerInfo{
				Name:    "linkding-mcp",
				Version: "1.0.0",
			},
			Capabilities: map[string]any{
				"tools": map[string]any{},
			},
		},
	}
}

func (s *Server) handleToolsList(request MCPRequest) MCPResponse {
	tools := []map[string]any{
		{
			"name":        "search_bookmarks",
			"description": "Search bookmarks in Linkding",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"query": map[string]any{
						"type":        "string",
						"description": "Search query",
					},
					"limit": map[string]any{
						"type":        "number",
						"description": "Maximum number of results (default: 20)",
						"default":     20,
					},
				},
			},
		},
		{
			"name":        "create_bookmark",
			"description": "Create a new bookmark",
			"inputSchema": map[string]any{
				"type": "object",
				"properties": map[string]any{
					"url": map[string]any{
						"type":        "string",
						"description": "URL to bookmark",
					},
					"title": map[string]any{
						"type":        "string",
						"description": "Bookmark title",
					},
					"description": map[string]any{
						"type":        "string",
						"description": "Bookmark description",
					},
					"tags": map[string]any{
						"type":        "array",
						"items":       map[string]any{"type": "string"},
						"description": "List of tags",
					},
				},
				"required": []string{"url"},
			},
		},
	}

	return MCPResponse{
		Jsonrpc: "2.0",
		ID:      request.ID,
		Result: map[string]any{
			"tools": tools,
		},
	}
}

type ToolCallParams struct {
	Name      string         `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

type ToolResult struct {
	Content []ToolContent `json:"content"`
	IsError bool          `json:"isError,omitempty"`
}

type ToolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

func (s *Server) handleToolsCall(ctx context.Context, request MCPRequest) MCPResponse {
	var params ToolCallParams
	paramsBytes, err := json.Marshal(request.Params)
	if err != nil {
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	if err := json.Unmarshal(paramsBytes, &params); err != nil {
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32602,
				Message: "Invalid params",
			},
		}
	}

	switch params.Name {
	case "search_bookmarks":
		return s.handleSearchBookmarks(ctx, request.ID, params.Arguments)
	case "create_bookmark":
		return s.handleCreateBookmark(ctx, request.ID, params.Arguments)
	default:
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      request.ID,
			Error: &MCPError{
				Code:    -32601,
				Message: "Tool not found",
			},
		}
	}
}

func (s *Server) handleSearchBookmarks(ctx context.Context, id any, args map[string]any) MCPResponse {
	query, _ := args["query"].(string)
	limit := 20
	if l, ok := args["limit"].(float64); ok {
		limit = int(l)
	}

	bookmarks, err := s.linkdingClient.GetBookmarks(ctx, limit, 0, query)
	if err != nil {
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      id,
			Result: ToolResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Failed to search bookmarks: %v", err),
					},
				},
			},
		}
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

	return MCPResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: result,
				},
			},
		},
	}
}

func (s *Server) handleCreateBookmark(ctx context.Context, id any, args map[string]any) MCPResponse {
	url, ok := args["url"].(string)
	if !ok || url == "" {
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      id,
			Result: ToolResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: "URL is required",
					},
				},
			},
		}
	}

	req := linkding.CreateBookmarkRequest{
		URL: url,
	}

	if title, ok := args["title"].(string); ok {
		req.Title = title
	}

	if description, ok := args["description"].(string); ok {
		req.Description = description
	}

	if tags, ok := args["tags"].([]interface{}); ok {
		for _, tag := range tags {
			if tagStr, ok := tag.(string); ok {
				req.TagNames = append(req.TagNames, tagStr)
			}
		}
	}

	bookmark, err := s.linkdingClient.CreateBookmark(ctx, req)
	if err != nil {
		return MCPResponse{
			Jsonrpc: "2.0",
			ID:      id,
			Result: ToolResult{
				IsError: true,
				Content: []ToolContent{
					{
						Type: "text",
						Text: fmt.Sprintf("Failed to create bookmark: %v", err),
					},
				},
			},
		}
	}

	result := fmt.Sprintf("✅ Bookmark created successfully!\n\n• **%s**\n  URL: %s\n  ID: %d",
		bookmark.Title, bookmark.URL, bookmark.ID)
	if bookmark.Description != "" {
		result += fmt.Sprintf("\n  Description: %s", bookmark.Description)
	}
	if len(bookmark.TagNames) > 0 {
		result += fmt.Sprintf("\n  Tags: %v", bookmark.TagNames)
	}

	return MCPResponse{
		Jsonrpc: "2.0",
		ID:      id,
		Result: ToolResult{
			Content: []ToolContent{
				{
					Type: "text",
					Text: result,
				},
			},
		},
	}
}
