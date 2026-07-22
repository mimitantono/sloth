package tools_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/http/backend/model"
	"github.com/slok/sloth/internal/http/mcp/tools"
	"github.com/slok/sloth/internal/http/mcp/tools/toolsmock"
	"github.com/slok/sloth/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewGetSLOBurnedBudgetRangeTool(t *testing.T) {
	ts1 := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 5, 25, 11, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		input   tools.GetSLOBurnedBudgetRangeToolInput
		mock    func(m *toolsmock.BurnedBudgetRangeLister)
		expErr  bool
		expResp tools.GetSLOBurnedBudgetRangeToolOutput
	}{
		"It should map the backend request and response.": {
			input: tools.GetSLOBurnedBudgetRangeToolInput{SLOID: "slo-id", RangeType: string(backendapp.BudgetRangeTypeMonthly)},
			mock: func(m *toolsmock.BurnedBudgetRangeLister) {
				expReq := backendapp.ListBurnedBudgetRangeRequest{SLOID: "slo-id", BudgetRangeType: backendapp.BudgetRangeTypeMonthly}
				m.On("ListBurnedBudgetRange", mock.Anything, expReq).Once().Return(&backendapp.ListBurnedBudgetRangeResponse{
					CurrentBurnedValuePercent:         87.1,
					CurrentExpectedBurnedValuePercent: 92.2,
					RealBurnedDataPoints:              []model.DataPoint{{TS: ts1, Value: 95.5}, {TS: ts2, Missing: true}},
					PerfectBurnedDataPoints:           []model.DataPoint{{TS: ts1, Value: 98.2}, {TS: ts2, Value: 97.1}},
				}, nil)
			},
			expResp: tools.GetSLOBurnedBudgetRangeToolOutput{
				CurrentBurnedValuePercent:         87.1,
				CurrentExpectedBurnedValuePercent: 92.2,
				StartTS:                           ts1.Format(time.RFC3339),
				Step:                              "1h0m0s",
				RealSeries:                        "95.5,x",
				PerfectSeries:                     "98.2,97.1",
			},
		},
		"Having a backend error should fail.": {
			input: tools.GetSLOBurnedBudgetRangeToolInput{SLOID: "slo-id", RangeType: string(backendapp.BudgetRangeTypeWeekly)},
			mock: func(m *toolsmock.BurnedBudgetRangeLister) {
				m.On("ListBurnedBudgetRange", mock.Anything, backendapp.ListBurnedBudgetRangeRequest{SLOID: "slo-id", BudgetRangeType: backendapp.BudgetRangeTypeWeekly}).Once().Return(nil, fmt.Errorf("something wrong"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := toolsmock.NewBurnedBudgetRangeLister(t)
			if test.mock != nil {
				test.mock(m)
			}

			tool, handler := tools.NewGetSLOBurnedBudgetRangeTool(m, log.Noop)
			require.NotNil(t, tool)
			assert.Equal(t, "get_slo_burned_budget_range", tool.Name)
			result, gotResp, err := handler(context.Background(), nil, test.input)
			assert.Nil(t, result)

			if test.expErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
			}

			assert.Equal(t, test.expResp, gotResp)
		})
	}
}
