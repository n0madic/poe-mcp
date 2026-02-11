package main

import (
	"context"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/types"
)

// QueryBotArgs defines the input schema for the query_bot tool.
type QueryBotArgs struct {
	Bot         string   `json:"bot" jsonschema:"Bot name on Poe.com (e.g. GPT-4o, Claude-4.5-Sonnet, Gemini-2.5-Pro)"`
	Message     string   `json:"message" jsonschema:"User message to send to the bot"`
	Temperature *float64 `json:"temperature,omitempty" jsonschema:"Sampling temperature (0.0-2.0)"`
}

func registerQueryBot(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_bot",
		Description: "Send a message to any Poe.com bot and get the full response",
	}, handleQueryBot)
}

func handleQueryBot(ctx context.Context, req *mcp.CallToolRequest, args QueryBotArgs) (*mcp.CallToolResult, any, error) {
	messages := []types.ProtocolMessage{
		{Role: "user", Content: args.Message},
	}

	queryReq := &types.QueryRequest{
		BaseRequest: types.BaseRequest{
			Version: types.ProtocolVersion,
			Type:    types.RequestTypeQuery,
		},
		Query:       messages,
		Temperature: args.Temperature,
	}

	response, err := client.GetFinalResponse(ctx, queryReq, args.Bot, apiKey, nil)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Error querying bot %q: %v", args.Bot, err)},
			},
			IsError: true,
		}, nil, nil
	}

	if response == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("Bot %q returned an empty response", args.Bot)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: response},
		},
	}, nil, nil
}
