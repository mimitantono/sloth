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

func TestNewGetSLOTool(t *testing.T) {
	tests := map[string]struct {
		input   tools.GetSLOToolInput
		mock    func(m *toolsmock.SLOGetter)
		expErr  bool
		expResp tools.GetSLOToolOutput
	}{
		"It should map the backend request and response.": {
			input: tools.GetSLOToolInput{SLOID: "slo-id"},
			mock: func(m *toolsmock.SLOGetter) {
				m.On("GetSLO", mock.Anything, backendapp.GetSLORequest{SLOID: "slo-id"}).Once().Return(&backendapp.GetSLOResponse{
					SLO: backendapp.RealTimeSLODetails{
						SLO: model.SLO{
							ID:             "slo-id",
							SlothID:        "sloth-slo-id",
							Name:           "availability",
							ServiceID:      "checkout",
							Objective:      99.9,
							PeriodDuration: 30 * 24 * time.Hour,
							IsGrouped:      true,
							GroupLabels:    map[string]string{"region": "eu-west-1"},
						},
						Budget: model.SLOBudgetDetails{
							BurningBudgetPercent:      123.4,
							BurnedBudgetWindowPercent: 77.7,
						},
						Alerts: model.SLOAlerts{
							FiringPage:    &model.Alert{Name: "PageAlert"},
							FiringWarning: &model.Alert{Name: "WarnAlert"},
						},
					},
				}, nil)
			},
			expResp: tools.GetSLOToolOutput{
				SLO: tools.ListSLOsToolOutputItem{
					ID:                        "slo-id",
					SlothID:                   "sloth-slo-id",
					Name:                      "availability",
					ServiceID:                 "checkout",
					Objective:                 99.9,
					Period:                    "720h0m0s",
					IsGrouped:                 true,
					GroupLabels:               map[string]string{"region": "eu-west-1"},
					BurningBudgetPercent:      123.4,
					BurnedBudgetWindowPercent: 77.7,
					HasPageAlert:              true,
					PageAlertName:             "PageAlert",
					HasWarningAlert:           true,
					WarningAlertName:          "WarnAlert",
				},
			},
		},
		"Having a backend error should fail.": {
			input: tools.GetSLOToolInput{SLOID: "slo-id"},
			mock: func(m *toolsmock.SLOGetter) {
				m.On("GetSLO", mock.Anything, backendapp.GetSLORequest{SLOID: "slo-id"}).Once().Return(nil, fmt.Errorf("something wrong"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := toolsmock.NewSLOGetter(t)
			if test.mock != nil {
				test.mock(m)
			}

			tool, handler := tools.NewGetSLOTool(m, log.Noop)
			require.NotNil(t, tool)
			assert.Equal(t, "get_slo", tool.Name)
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
