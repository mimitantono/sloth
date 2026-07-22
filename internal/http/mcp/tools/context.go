package tools

import (
	"context"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/slok/sloth/internal/info"
	"github.com/slok/sloth/internal/log"
)

func NewContextTool(logger log.Logger) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[ContextToolInput, ContextToolOutput]) {
	if logger == nil {
		logger = log.Noop
	}

	return &sdkmcp.Tool{
		Name:        "context",
		Description: "Get context about Sloth and its SLO framework. Returns the running Sloth version and a description of what Sloth does.",
		Annotations: &sdkmcp.ToolAnnotations{
			ReadOnlyHint: true,
		},
	}, newContextToolHandler(logger)
}

type ContextToolInput struct{}

type ContextToolOutput struct {
	Version     string `json:"version" jsonschema:"the running Sloth version"`
	Description string `json:"description" jsonschema:"what Sloth is and what it does"`
}

func newContextToolHandler(logger log.Logger) sdkmcp.ToolHandlerFor[ContextToolInput, ContextToolOutput] {
	return func(_ context.Context, _ *sdkmcp.CallToolRequest, input ContextToolInput) (*sdkmcp.CallToolResult, ContextToolOutput, error) {
		logger.WithValues(log.Kv{"input": fmt.Sprintf("%+v", input)}).Debugf("MCP tool called")

		return nil, ContextToolOutput{
			Version:     info.Version,
			Description: "Sloth is a Prometheus SLO framework that helps teams define service level objectives and creates a uniform, standardized layer of low-level Prometheus rules to implement SLOs, including the recording and alerting rules required to measure them.",
		}, nil
	}
}
