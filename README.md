# Poe-MCP

MCP (Model Context Protocol) server and CLI tool that exposes [Poe](https://poe.com) bots and the model catalog. Use it as:
- **MCP Server**: Integrate with MCP-compatible clients (Claude Code, Cursor, etc.) to query bots and search models
- **CLI Tool**: Run directly from the terminal for quick bot queries and model searches

## Usage

### MCP Server Mode

Run `poe-mcp` without arguments to start the MCP server:

```bash
POE_API_KEY=<key> poe-mcp
```

Then configure your MCP client (see [Configuration](#configuration) below).

## MCP Tools

### `query_bot`

Send a message to any Poe bot and get the full response.

| Parameter     | Type   | Required | Description                                    |
|---------------|--------|----------|------------------------------------------------|
| `bot`         | string | yes      | Bot name on Poe (e.g. GPT-4o, Claude-4.5-Sonnet) |
| `message`     | string | yes      | User message to send to the bot                |
| `temperature` | float  | no       | Sampling temperature (0.0–2.0)                 |

### `search_models`

Search and filter the Poe model catalog.

| Parameter  | Type   | Required | Description                                      |
|------------|--------|----------|--------------------------------------------------|
| `query`    | string | no       | Case-insensitive substring match on ID, name, description, owner |
| `owned_by` | string | no       | Filter by owner/provider (e.g. OpenAI, Anthropic) |
| `modality` | string | no       | Filter by modality substring (e.g. text, image)  |

### CLI Mode

Run `poe-mcp` with subcommands for direct terminal access:

**Search models** (no API key required):
```bash
# Search all models
poe-mcp search "GPT-4o"

# Filter by owner
poe-mcp search --owner OpenAI

# Filter by modality
poe-mcp search --modality image

# Combine filters and query
poe-mcp search --owner Google --modality text "pro"
```

**Query a bot** (requires `POE_API_KEY`):
```bash
export POE_API_KEY=<key>
# Basic query
poe-mcp query GPT-4o "What is Go?"

# With temperature setting
poe-mcp query -t 0.7 Claude-4.5-Sonnet "Explain monads"
poe-mcp query --temperature 0.9 GPT-5.2-Pro "Write a poem"
```

## Installation

```bash
go install github.com/n0madic/poe-mcp@latest
```

Or build from source:

```bash
git clone https://github.com/n0madic/poe-mcp.git
cd poe-mcp
go build .
```

## Configuration

### Claude Code

Add to your Claude Code MCP settings (`~/.claude/claude_desktop_config.json` or project `.mcp.json`):

```json
{
  "mcpServers": {
    "poe": {
      "command": "poe-mcp",
      "env": {
        "POE_API_KEY": "your-poe-api-key"
      }
    }
  }
}
```

### Cursor / Other MCP Clients

Configure the MCP server with:
- **Command:** `poe-mcp`
- **Environment:** `POE_API_KEY=your-poe-api-key`
- **Transport:** stdio

## Environment Variables

| Variable      | Required | Description                              |
|---------------|----------|------------------------------------------|
| `POE_API_KEY` | For MCP server mode and `query` CLI command | Poe API key for bot queries. Not required for `search` CLI command. |

## Getting a Poe API Key

1. Go to [poe.com](https://poe.com/api/keys)
2. Generate or copy your API key

## License

MIT
