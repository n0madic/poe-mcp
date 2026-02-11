package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/models"
	"github.com/n0madic/go-poe/types"
)

// runCLI handles CLI mode subcommands (search, query).
func runCLI(args []string) error {
	if len(args) == 0 {
		printHelp()
		return nil
	}

	subcommand := args[0]
	switch subcommand {
	case "help", "--help", "-h":
		printHelp()
		return nil
	case "search":
		return runSearch(args[1:])
	case "query":
		return runQuery(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "Error: unknown subcommand %q\n\n", subcommand)
		printHelp()
		return fmt.Errorf("unknown subcommand")
	}
}

// printHelp displays usage information.
func printHelp() {
	fmt.Println(`poe-mcp - Poe.com MCP Server and CLI Tool

USAGE:
    poe-mcp              Start MCP server (stdio transport)
    poe-mcp search       Search and filter Poe model catalog
    poe-mcp query        Query a Poe bot and stream response

COMMANDS:
    search [flags] [query]
        Search the Poe model catalog (no API key required)

        Flags:
          --owner string      Filter by owner/provider (e.g., OpenAI, Anthropic)
          --modality string   Filter by modality (e.g., text, image)

        Examples:
          poe-mcp search "GPT-4o"
          poe-mcp search --owner OpenAI
          poe-mcp search --owner Google --modality text "pro"

    query [flags] <bot> <message>
        Query a Poe bot and stream the response (requires POE_API_KEY)

        Flags:
          -t, --temperature float   Sampling temperature 0.0-2.0 (default: 0.7)

        Examples:
          POE_API_KEY=<key> poe-mcp query GPT-4o "What is Go?"
          POE_API_KEY=<key> poe-mcp query -t 0.9 Claude-4.5-Sonnet "Explain monads"

ENVIRONMENT VARIABLES:
    POE_API_KEY    Required for MCP server mode and 'query' command
                   Not required for 'search' command`)
}

// runSearch handles the 'search' subcommand.
func runSearch(args []string) error {
	fs := flag.NewFlagSet("search", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Println(`Usage: poe-mcp search [flags] [query]

Search and filter the Poe model catalog (no API key required).

FLAGS:
  --owner string      Filter by owner/provider (e.g., OpenAI, Anthropic)
  --modality string   Filter by modality (e.g., text, image)

EXAMPLES:
  poe-mcp search "GPT-4o"
  poe-mcp search --owner OpenAI
  poe-mcp search --owner Google --modality text "pro"`)
	}
	owner := fs.String("owner", "", "Filter by owner/provider (e.g., OpenAI, Anthropic)")
	modality := fs.String("modality", "", "Filter by modality (e.g., text, image)")

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil // Help was printed, exit cleanly
		}
		return err
	}

	// Remaining positional args form the query string
	query := strings.Join(fs.Args(), " ")

	// Fetch models from the public API (no API key needed)
	ctx := context.Background()
	all, err := models.Fetch(ctx, nil)
	if err != nil {
		return fmt.Errorf("error fetching models: %w", err)
	}

	// Build search args and filter
	searchArgs := SearchModelsArgs{
		Query:    query,
		OwnedBy:  *owner,
		Modality: *modality,
	}
	matched := filterModels(all, searchArgs)

	if len(matched) == 0 {
		fmt.Println("No models found matching the given criteria.")
		return nil
	}

	// Print formatted results
	fmt.Print(formatModels(matched))
	return nil
}

// runQuery handles the 'query' subcommand.
func runQuery(args []string) error {
	fs := flag.NewFlagSet("query", flag.ContinueOnError)
	fs.Usage = func() {
		fmt.Println(`Usage: poe-mcp query [flags] <bot> <message>

Query a Poe bot and stream the response (requires POE_API_KEY).

FLAGS:
  -t, --temperature float   Sampling temperature 0.0-2.0 (default: 0.7)

EXAMPLES:
  POE_API_KEY=<key> poe-mcp query GPT-4o "What is Go?"
  POE_API_KEY=<key> poe-mcp query -t 0.9 Claude-4.5-Sonnet "Explain monads"`)
	}
	temperature := fs.Float64("t", 0.7, "Sampling temperature (0.0-2.0)")
	fs.Float64("temperature", 0.7, "Sampling temperature (0.0-2.0)") // Alias

	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return nil // Help was printed, exit cleanly
		}
		return err
	}

	// Two positional args required: <bot> <message>
	positional := fs.Args()
	if len(positional) < 2 {
		return fmt.Errorf("usage: query [-t temperature] <bot> <message>")
	}

	bot := positional[0]
	message := strings.Join(positional[1:], " ")

	// POE_API_KEY is required for querying
	apiKey := os.Getenv("POE_API_KEY")
	if apiKey == "" {
		return fmt.Errorf("POE_API_KEY environment variable is required for query command")
	}

	// Construct the message
	messages := []types.ProtocolMessage{
		{Role: "user", Content: message},
	}

	// Stream the response
	ctx := context.Background()
	opts := &client.StreamRequestOptions{
		APIKey: apiKey,
	}

	// Build query request with temperature
	req := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:       messages,
		Temperature: temperature,
	}

	ch := client.StreamRequest(ctx, req, bot, opts)

	// Print each chunk as it arrives
	for chunk := range ch {
		// Skip metadata and suggested replies
		if chunk.RawResponse != nil {
			if _, ok := chunk.RawResponse.(*types.MetaResponse); ok {
				continue
			}
		}
		if chunk.IsSuggestedReply {
			continue
		}

		// Print the text chunk
		if chunk.Text != "" {
			fmt.Print(chunk.Text)
		}
	}

	fmt.Println() // Newline at the end
	return nil
}
