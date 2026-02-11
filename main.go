package main

import (
	"context"
	"log"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var apiKey string

func main() {
	// If subcommand provided, run CLI mode.
	if len(os.Args) > 1 {
		if err := runCLI(os.Args[1:]); err != nil {
			log.Fatal(err)
		}
		return
	}

	// Otherwise, run as MCP server (current behavior).
	apiKey = os.Getenv("POE_API_KEY")
	if apiKey == "" {
		log.Fatal("POE_API_KEY environment variable is required")
	}

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "poe-mcp",
			Title:   "Poe.com MCP Server",
			Version: "1.0.0",
		},
		nil,
	)

	registerQueryBot(server)
	registerSearchModels(server)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
