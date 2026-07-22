package tools

import (
	"context"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/log"
)

type SLOGetter interface {
	GetSLO(ctx context.Context, req backendapp.GetSLORequest) (*backendapp.GetSLOResponse, error)
}

func NewGetSLOTool(app SLOGetter, logger log.Logger) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[GetSLOToolInput, GetSLOToolOutput]) {
	if logger == nil {
		logger = log.Noop
	}

	return &sdkmcp.Tool{
		Name:        "get_slo",
		Description: "Get a single SLO with its metadata, current budget burn, consumed budget in the window, and firing alert status.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true},
	}, newGetSLOToolHandler(app, logger)
}

type GetSLOToolInput struct {
	SLOID string `json:"slo_id" jsonschema:"required,The SLO ID to retrieve"`
}

type GetSLOToolOutput struct {
	SLO ListSLOsToolOutputItem `json:"slo" jsonschema:"the requested SLO with its current status"`
}

func newGetSLOToolHandler(app SLOGetter, logger log.Logger) sdkmcp.ToolHandlerFor[GetSLOToolInput, GetSLOToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetSLOToolInput) (*sdkmcp.CallToolResult, GetSLOToolOutput, error) {
		logger.WithValues(log.Kv{"input": fmt.Sprintf("%+v", input)}).Debugf("MCP tool called")

		resp, err := app.GetSLO(ctx, backendapp.GetSLORequest{SLOID: input.SLOID})
		if err != nil {
			return nil, GetSLOToolOutput{}, err
		}

		return nil, GetSLOToolOutput{SLO: mapRealTimeSLOToToolOutputItem(resp.SLO)}, nil
	}
}
