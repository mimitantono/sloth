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

func TestNewListSLOsTool(t *testing.T) {
	tests := map[string]struct {
		input   tools.ListSLOsToolInput
		mock    func(m *toolsmock.SLOLister)
		expErr  bool
		expResp tools.ListSLOsToolOutput
	}{
		"It should map the backend request and response.": {
			input: tools.ListSLOsToolInput{
				ServiceID:                   "checkout",
				Search:                      "availability",
				AlertFiring:                 true,
				PeriodBudgetConsumed:        true,
				CurrentBurningBudgetOver100: true,
				Size:                        100,
				Sort:                        string(backendapp.SLOListSortModeCurrentBurningBudgetDesc),
				Cursor:                      "cursor-1",
			},
			mock: func(m *toolsmock.SLOLister) {
				expReq := backendapp.ListSLOsRequest{
					FilterServiceID:                   "checkout",
					FilterSearchInput:                 "availability",
					FilterAlertFiring:                 true,
					FilterPeriodBudgetConsumed:        true,
					FilterCurrentBurningBudgetOver100: true,
					PageSize:                          100,
					SortMode:                          backendapp.SLOListSortModeCurrentBurningBudgetDesc,
					Cursor:                            "cursor-1",
				}
				m.On("ListSLOs", mock.Anything, expReq).Once().Return(&backendapp.ListSLOsResponse{
					SLOs: []backendapp.RealTimeSLODetails{{
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
					}},
					PaginationCursors: backendapp.PaginationCursors{
						NextCursor:  "next-1",
						PrevCursor:  "prev-1",
						HasNext:     true,
						HasPrevious: true,
					},
				}, nil)
			},
			expResp: tools.ListSLOsToolOutput{
				SLOs: []tools.ListSLOsToolOutputItem{{
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
				}},
				Pagination: tools.ListSLOsToolOutputPagination{
					NextCursor:  "next-1",
					PrevCursor:  "prev-1",
					HasNext:     true,
					HasPrevious: true,
				},
			},
		},
		"It should default page size to 100.": {
			mock: func(m *toolsmock.SLOLister) {
				m.On("ListSLOs", mock.Anything, mock.MatchedBy(func(req backendapp.ListSLOsRequest) bool { return req.PageSize == 100 })).Once().Return(&backendapp.ListSLOsResponse{}, nil)
			},
			expResp: tools.ListSLOsToolOutput{SLOs: []tools.ListSLOsToolOutputItem{}, Pagination: tools.ListSLOsToolOutputPagination{HasNext: false, HasPrevious: false}},
		},
		"Having a backend error should fail.": {
			mock: func(m *toolsmock.SLOLister) {
				m.On("ListSLOs", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("something wrong"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := toolsmock.NewSLOLister(t)
			if test.mock != nil {
				test.mock(m)
			}

			tool, handler := tools.NewListSLOsTool(m, log.Noop)
			require.NotNil(t, tool)
			assert.Equal(t, "list_slos", tool.Name)

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
