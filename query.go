package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/n0madic/go-poe/client"
	"github.com/n0madic/go-poe/types"
)

// FileInput describes a file to upload and attach to the message.
type FileInput struct {
	URL     string `json:"url,omitempty" jsonschema:"URL of the file to upload (mutually exclusive with content)"`
	Content string `json:"content,omitempty" jsonschema:"Base64-encoded file content (mutually exclusive with url)"`
	Name    string `json:"name" jsonschema:"Filename (e.g. document.pdf, image.png)"`
}

// QueryBotArgs defines the input schema for the query_bot tool.
type QueryBotArgs struct {
	Bot         string      `json:"bot" jsonschema:"Bot name on Poe.com (e.g. GPT-4o, Claude-4.5-Sonnet, Gemini-2.5-Pro)"`
	Message     string      `json:"message" jsonschema:"User message to send to the bot"`
	Files       []FileInput `json:"files,omitempty" jsonschema:"Files to attach to the message"`
	Temperature *float64    `json:"temperature,omitempty" jsonschema:"Sampling temperature (0.0-2.0)"`
}

func registerQueryBot(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "query_bot",
		Description: "Send a message to any Poe.com bot and get the full response",
	}, handleQueryBot)
}

// uploadFiles uploads each FileInput and returns the resulting attachments.
func uploadFiles(ctx context.Context, files []FileInput, key string) ([]types.Attachment, error) {
	var attachments []types.Attachment
	for _, f := range files {
		if f.Name == "" {
			return nil, fmt.Errorf("file name is required")
		}
		if f.URL == "" && f.Content == "" {
			return nil, fmt.Errorf("file %q: either url or content is required", f.Name)
		}

		opts := &client.UploadFileOptions{
			FileName: f.Name,
			APIKey:   key,
		}

		if f.URL != "" {
			opts.FileURL = f.URL
		} else {
			data, err := base64.StdEncoding.DecodeString(f.Content)
			if err != nil {
				return nil, fmt.Errorf("file %q: invalid base64 content: %w", f.Name, err)
			}
			opts.File = bytes.NewReader(data)
		}

		att, err := client.UploadFile(ctx, opts)
		if err != nil {
			return nil, fmt.Errorf("file %q: upload failed: %w", f.Name, err)
		}
		attachments = append(attachments, *att)
	}
	return attachments, nil
}

func handleQueryBot(ctx context.Context, req *mcp.CallToolRequest, args QueryBotArgs) (*mcp.CallToolResult, any, error) {
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
