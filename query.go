package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/types"
)

// QueryBotArgs defines the input schema for the query_bot tool.
type QueryBotArgs struct {
	Bot         string   `json:"bot" jsonschema:"Bot name on Poe.com (e.g. GPT-4o, Claude-4.5-Sonnet, Gemini-2.5-Pro)"`
	Message     string   `json:"message" jsonschema:"User message to send to the bot"`
	Files       []string `json:"files,omitempty" jsonschema:"Files to attach (local paths or URLs)"`
	Temperature *float64 `json:"temperature,omitempty" jsonschema:"Sampling temperature (0.0-2.0)"`
}

func registerQueryBot(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_bot",
		Description: "Send a message to any Poe.com bot and get the full response",
	}, handleQueryBot)
}

// uploadFiles uploads each file path or URL and returns the resulting attachments.
// Strings starting with http:// or https:// are treated as URLs; everything else
// is treated as a local file path.
func uploadFiles(ctx context.Context, files []string, key string) ([]types.Attachment, error) {
	var attachments []types.Attachment
	for _, path := range files {
		att, err := uploadSingleFile(ctx, path, key)
		if err != nil {
			return nil, err
		}
		attachments = append(attachments, *att)
	}
	return attachments, nil
}

// uploadSingleFile uploads a single file (local path or URL) and returns the attachment.
func uploadSingleFile(ctx context.Context, path, key string) (*types.Attachment, error) {
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		name := filepath.Base(path)
		if name == "" || name == "." || name == "/" {
			name = "file"
		}
		return client.UploadFile(ctx, &client.UploadFileOptions{
			FileURL:  path,
			FileName: name,
			APIKey:   key,
		})
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("file %q: %w", path, err)
	}
	defer f.Close()

	return client.UploadFile(ctx, &client.UploadFileOptions{
		File:     f,
		FileName: filepath.Base(path),
		APIKey:   key,
	})
}

func handleQueryBot(ctx context.Context, req *mcp.CallToolRequest, args QueryBotArgs) (*mcp.CallToolResult, any, error) {
	if apiKey == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "POE_API_KEY environment variable is required"},
			},
			IsError: true,
		}, nil, nil
	}

	var attachments []types.Attachment
	if len(args.Files) > 0 {
		var err error
		attachments, err = uploadFiles(ctx, args.Files, apiKey)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: fmt.Sprintf("Error uploading files: %v", err)},
				},
				IsError: true,
			}, nil, nil
		}
	}

	messages := []types.ProtocolMessage{
		{Role: "user", Content: args.Message, Attachments: attachments},
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
