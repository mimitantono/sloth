package tools

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/http/backend/model"
	"github.com/slok/sloth/internal/log"
)

type SLIAvailabilityRangeLister interface {
	ListSLIAvailabilityRange(ctx context.Context, req backendapp.ListSLIAvailabilityRangeRequest) (*backendapp.ListSLIAvailabilityRangeResponse, error)
}

func NewGetSLOSLIAvailabilityRangeTool(app SLIAvailabilityRangeLister, logger log.Logger) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[GetSLOSLIAvailabilityRangeToolInput, GetSLOSLIAvailabilityRangeToolOutput]) {
	if logger == nil {
		logger = log.Noop
	}

	return &sdkmcp.Tool{
		Name:        "get_slo_sli_availability_range",
		Description: "Get SLI availability evolution for an SLO over a time range. Returns a compressed availability series where each comma-separated entry advances by one step from start_ts, and x means the value is missing at that step.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true},
	}, newGetSLOSLIAvailabilityRangeToolHandler(app, logger)
}

type GetSLOSLIAvailabilityRangeToolInput struct {
	SLOID string `json:"slo_id" jsonschema:"required,The SLO ID to retrieve"`
	From  string `json:"from" jsonschema:"required,The RFC3339 start timestamp of the availability range"`
	To    string `json:"to,omitempty" jsonschema:"Optional RFC3339 end timestamp of the availability range. If omitted, now is used"`
}

type GetSLOSLIAvailabilityRangeToolOutput struct {
	StartTS            string `json:"start_ts" jsonschema:"the RFC3339 timestamp of the first point in the series"`
	Step               string `json:"step" jsonschema:"the fixed duration between series points"`
	AvailabilitySeries string `json:"availability_series" jsonschema:"comma-separated SLI availability percentage values from start_ts advancing by step. Use x when a value is missing"`
}

func newGetSLOSLIAvailabilityRangeToolHandler(app SLIAvailabilityRangeLister, logger log.Logger) sdkmcp.ToolHandlerFor[GetSLOSLIAvailabilityRangeToolInput, GetSLOSLIAvailabilityRangeToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetSLOSLIAvailabilityRangeToolInput) (*sdkmcp.CallToolResult, GetSLOSLIAvailabilityRangeToolOutput, error) {
		logger.WithValues(log.Kv{"input": fmt.Sprintf("%+v", input)}).Debugf("MCP tool called")

		from, err := time.Parse(time.RFC3339, input.From)
		if err != nil {
			return nil, GetSLOSLIAvailabilityRangeToolOutput{}, fmt.Errorf("invalid from time: %w", err)
		}

		to := time.Time{}
		if input.To != "" {
			to, err = time.Parse(time.RFC3339, input.To)
			if err != nil {
				return nil, GetSLOSLIAvailabilityRangeToolOutput{}, fmt.Errorf("invalid to time: %w", err)
			}
		}

		resp, err := app.ListSLIAvailabilityRange(ctx, backendapp.ListSLIAvailabilityRangeRequest{
			SLOID: input.SLOID,
			From:  from,
			To:    to,
		})
		if err != nil {
			return nil, GetSLOSLIAvailabilityRangeToolOutput{}, err
		}

		startTS, step := getAvailabilitySeriesMeta(resp.AvailabilityDataPoints)

		return nil, GetSLOSLIAvailabilityRangeToolOutput{
			StartTS:            startTS,
			Step:               step,
			AvailabilitySeries: compressAvailabilityDataPoints(resp.AvailabilityDataPoints),
		}, nil
	}
}

func getAvailabilitySeriesMeta(dps []model.DataPoint) (startTS, step string) {
	if len(dps) == 0 {
		return "", ""
	}

	startTS = dps[0].TS.Format(time.RFC3339)
	if len(dps) < 2 {
		return startTS, "0s"
	}

	return startTS, dps[1].TS.Sub(dps[0].TS).String()
}

func compressAvailabilityDataPoints(dps []model.DataPoint) string {
	parts := make([]string, 0, len(dps))
	for _, dp := range dps {
		if dp.Missing {
			parts = append(parts, "x")
			continue
		}

		parts = append(parts, strconv.FormatFloat(dp.Value, 'f', -1, 64))
	}

	return strings.Join(parts, ",")
}
