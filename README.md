# Linkding MCP Server

A Model Context Protocol (MCP) server that provides seamless integration with [Linkding](https://linkding.link/), a self-hosted bookmark manager. This server enables AI assistants like Claude to interact with your Linkding bookmarks through standardized MCP tools.

## Features

- 🔍 **Search Bookmarks** - Find bookmarks by query with customizable limits
- ➕ **Create Bookmarks** - Add new bookmarks with titles, descriptions, and tags
- 🏷️ **Manage Tags** - Retrieve and work with your bookmark tags
- 🚀 **Dual Mode Support** - Run via stdio (MCP standard) or HTTP server
- 🐳 **Docker Ready** - Multi-stage Alpine-based container
- 🔒 **Secure** - Runs as non-root user with minimal dependencies

## Tools Available

### `search_bookmarks`
Search through your Linkding bookmarks.

**Parameters:**
- `query` (string, optional): Search phrase to filter bookmarks
- `limit` (number, optional): Maximum results to return (default: 20)

### `create_bookmark` 
Create a new bookmark in Linkding.

**Parameters:**
- `url` (string, required): The URL to bookmark
- `title` (string, optional): Title for the bookmark
- `description` (string, optional): Description of the bookmark  
- `tags` (array of strings, optional): Tags to associate with the bookmark

### `get_tags`
Retrieve available tags from Linkding.

**Parameters:**
- `limit` (number, optional): Maximum number of tags to return (default: 50)

## Installation

### Prerequisites

- [Linkding](https://linkding.link/) instance running and accessible
- Linkding API token (generate in your Linkding admin panel)

### Option 1: Docker (Recommended)

```bash
# Pull the image
docker pull ghcr.io/chickenzord/linkding-mcp:latest

# Run in stdio mode (for MCP clients)
docker run --rm -i \
  -e LINKDING_URL="https://your-linkding.example.com" \
  -e LINKDING_API_TOKEN="your-api-token" \
  ghcr.io/chickenzord/linkding-mcp:latest stdio

# Run in HTTP mode  
docker run --rm -p 8080:8080 \
  -e LINKDING_URL="https://your-linkding.example.com" \
  -e LINKDING_API_TOKEN="your-api-token" \
  -e BIND_ADDR=":8080" \
  ghcr.io/chickenzord/linkding-mcp:latest http
```

### Option 2: Homebrew (macOS/Linux)

```bash
# Add the tap
brew tap chickenzord/tap

# Install linkding-mcp
brew install linkding-mcp

# Install latest development version
brew install --HEAD linkding-mcp
```

### Option 3: Go Install

```bash
go install github.com/chickenzord/linkding-mcp/cmd/linkding-mcp@latest
```

### Option 4: Build from Source

```bash
git clone https://github.com/chickenzord/linkding-mcp.git
cd linkding-mcp
go build -o linkding-mcp ./cmd/linkding-mcp
```

## Configuration

### Environment Variables

- `LINKDING_URL` (required): Your Linkding instance URL
- `LINKDING_API_TOKEN` (required): API token from your Linkding admin panel
- `BIND_ADDR` (optional): HTTP server bind address (default: ":8080")

### Getting Your Linkding API Token

1. Log into your Linkding instance
2. Go to Settings → Integrations
3. Generate a new API token
4. Copy the token for use with this MCP server

## MCP Client Configuration

### Claude Desktop

Add to your Claude Desktop configuration file:

**macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`  
**Windows**: `%APPDATA%/Claude/claude_desktop_config.json`

```json
{
  "mcpServers": {
    "linkding": {
      "command": "docker",
      "args": [
        "run", "--rm", "-i",
        "-e", "LINKDING_URL=https://your-linkding.example.com",
        "-e", "LINKDING_API_TOKEN=your-api-token",
        "ghcr.io/chickenzord/linkding-mcp:latest",
        "stdio"
      ]
    }
  }
}
```

Or if using a local binary:

```json
{
  "mcpServers": {
    "linkding": {
      "command": "linkding-mcp",
      "args": ["stdio"],
      "env": {
        "LINKDING_URL": "https://your-linkding.example.com",
        "LINKDING_API_TOKEN": "your-api-token"
      }
    }
  }
}
```

### VS Code with MCP Extension

```json
{
  "mcp.servers": [
    {
      "name": "linkding",
      "command": "linkding-mcp",
      "args": ["stdio"],
      "env": {
        "LINKDING_URL": "https://your-linkding.example.com", 
        "LINKDING_API_TOKEN": "your-api-token"
      }
    }
  ]
}
```

### Other MCP Clients

For any MCP-compatible client, use:
- **Command**: `linkding-mcp stdio` (or Docker equivalent)
- **Transport**: stdio
- **Environment**: Set `LINKDING_URL` and `LINKDING_API_TOKEN`

## Usage Examples

Once configured with an MCP client, you can use natural language commands:

- *"Search for bookmarks about golang"*
- *"Create a bookmark for https://example.com with title 'Example Site' and tags 'reference', 'tools'"*
- *"Show me all my tags"*
- *"Find bookmarks related to kubernetes"*

## Development

### Running Locally

```bash
# Set environment variables
export LINKDING_URL="https://your-linkding.example.com"
export LINKDING_API_TOKEN="your-api-token"

# Run in stdio mode
go run ./cmd/linkding-mcp stdio

# Run in HTTP mode  
go run ./cmd/linkding-mcp http
```

### Testing

```bash
# Run tests
go test ./...

# Test with a real Linkding instance
go run ./cmd/linkding-mcp stdio < test-requests.json
```

### Building

```bash
# Build binary
go build -o linkding-mcp ./cmd/linkding-mcp

# Build Docker image
docker build -t linkding-mcp .
```

## Docker Compose Example

```yaml
version: '3.8'
services:
  linkding-mcp:
    image: ghcr.io/chickenzord/linkding-mcp:latest
    environment:
      - LINKDING_URL=https://your-linkding.example.com
      - LINKDING_API_TOKEN=your-api-token
      - BIND_ADDR=:8080
    ports:
      - "8080:8080"
    command: ["http"]
```

## API Reference

The HTTP mode exposes MCP endpoints at:
- `POST /mcp/v1/initialize` - Initialize MCP session
- `POST /mcp/v1/tools/list` - List available tools  
- `POST /mcp/v1/tools/call` - Call a tool

See the [MCP specification](https://spec.modelcontextprotocol.io/) for detailed API documentation.

## Troubleshooting

### Common Issues

**Authentication Error**: Verify your `LINKDING_API_TOKEN` is correct and has proper permissions.

**Connection Error**: Ensure your `LINKDING_URL` is accessible and includes the protocol (https://).

**Permission Denied**: If running Docker, ensure the container can access your Linkding instance (check firewalls/networks).

### Debug Mode

Set `LINKDING_DEBUG=true` for verbose logging:

```bash
LINKDING_DEBUG=true linkding-mcp stdio
```

### Logs

Container logs:
```bash
docker logs <container-id>
```

## Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)  
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Related Projects

- [Linkding](https://linkding.link/) - The self-hosted bookmark manager
- [MCP Specification](https://spec.modelcontextprotocol.io/) - Model Context Protocol specification
- [Claude Desktop](https://claude.ai/desktop) - AI assistant with MCP support

## Support

- 🐛 **Bug Reports**: [GitHub Issues](https://github.com/chickenzord/linkding-mcp/issues)
- 💡 **Feature Requests**: [GitHub Discussions](https://github.com/chickenzord/linkding-mcp/discussions)
- 📖 **Documentation**: [Wiki](https://github.com/chickenzord/linkding-mcp/wiki)

---

Built with ❤️ for the MCP ecosystem
