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

type BurnedBudgetRangeLister interface {
	ListBurnedBudgetRange(ctx context.Context, req backendapp.ListBurnedBudgetRangeRequest) (*backendapp.ListBurnedBudgetRangeResponse, error)
}

func NewGetSLOBurnedBudgetRangeTool(app BurnedBudgetRangeLister, logger log.Logger) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[GetSLOBurnedBudgetRangeToolInput, GetSLOBurnedBudgetRangeToolOutput]) {
	if logger == nil {
		logger = log.Noop
	}

	return &sdkmcp.Tool{
		Name:        "get_slo_burned_budget_range",
		Description: "Get actual and expected burned budget evolution for an SLO over a standard time range. Values use a 0-100 normalized remaining-budget scale over the selected period, where 100 means the full budget is still available and 0 means the budget is exhausted. Returns current real and expected values plus compressed real and perfect series. Each comma-separated series entry advances by one step from start_ts, and x means the real value is missing at that step.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true},
	}, newGetSLOBurnedBudgetRangeToolHandler(app, logger)
}

type GetSLOBurnedBudgetRangeToolInput struct {
	SLOID     string `json:"slo_id" jsonschema:"required,The SLO ID to retrieve"`
	RangeType string `json:"range_type,omitempty" jsonschema:"Optional range type: monthly, weekly, quarterly, yearly"`
}

type GetSLOBurnedBudgetRangeToolOutput struct {
	CurrentBurnedValuePercent         float64 `json:"current_burned_value_percent" jsonschema:"the current real remaining budget percentage on a 0-100 scale for the selected period"`
	CurrentExpectedBurnedValuePercent float64 `json:"current_expected_burned_value_percent" jsonschema:"the current expected remaining budget percentage on a 0-100 scale for the selected period"`
	StartTS                           string  `json:"start_ts" jsonschema:"the RFC3339 timestamp of the first point in both series"`
	Step                              string  `json:"step" jsonschema:"the fixed duration between series points"`
	RealSeries                        string  `json:"real_series" jsonschema:"comma-separated real remaining budget values from start_ts advancing by step. Use x when a real point is missing"`
	PerfectSeries                     string  `json:"perfect_series" jsonschema:"comma-separated expected remaining budget values from start_ts advancing by step"`
}

func newGetSLOBurnedBudgetRangeToolHandler(app BurnedBudgetRangeLister, logger log.Logger) sdkmcp.ToolHandlerFor[GetSLOBurnedBudgetRangeToolInput, GetSLOBurnedBudgetRangeToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input GetSLOBurnedBudgetRangeToolInput) (*sdkmcp.CallToolResult, GetSLOBurnedBudgetRangeToolOutput, error) {
		logger.WithValues(log.Kv{"input": fmt.Sprintf("%+v", input)}).Debugf("MCP tool called")

		resp, err := app.ListBurnedBudgetRange(ctx, backendapp.ListBurnedBudgetRangeRequest{
			SLOID:           input.SLOID,
			BudgetRangeType: backendapp.BudgetRangeType(input.RangeType),
		})
		if err != nil {
			return nil, GetSLOBurnedBudgetRangeToolOutput{}, err
		}

		startTS, step := getSeriesMeta(resp.PerfectBurnedDataPoints)
		if startTS == "" {
			startTS, step = getSeriesMeta(resp.RealBurnedDataPoints)
		}

		return nil, GetSLOBurnedBudgetRangeToolOutput{
			CurrentBurnedValuePercent:         resp.CurrentBurnedValuePercent,
			CurrentExpectedBurnedValuePercent: resp.CurrentExpectedBurnedValuePercent,
			StartTS:                           startTS,
			Step:                              step,
			RealSeries:                        compressBurnedBudgetDataPoints(resp.RealBurnedDataPoints, true),
			PerfectSeries:                     compressBurnedBudgetDataPoints(resp.PerfectBurnedDataPoints, false),
		}, nil
	}
}

func getSeriesMeta(dps []model.DataPoint) (startTS, step string) {
	if len(dps) == 0 {
		return "", ""
	}

	startTS = dps[0].TS.Format(time.RFC3339)
	if len(dps) < 2 {
		return startTS, "0s"
	}

	return startTS, dps[1].TS.Sub(dps[0].TS).String()
}

func compressBurnedBudgetDataPoints(dps []model.DataPoint, allowMissing bool) string {
	parts := make([]string, 0, len(dps))
	for _, dp := range dps {
		if allowMissing && dp.Missing {
			parts = append(parts, "x")
			continue
		}

		parts = append(parts, strconv.FormatFloat(dp.Value, 'f', -1, 64))
	}

	return strings.Join(parts, ",")
}
