package tools_test

import (
	"context"
	"fmt"
	"testing"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/http/backend/model"
	"github.com/slok/sloth/internal/http/mcp/tools"
	"github.com/slok/sloth/internal/http/mcp/tools/toolsmock"
	"github.com/slok/sloth/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewListServicesTool(t *testing.T) {
	tests := map[string]struct {
		input   tools.ListServicesToolInput
		mock    func(m *toolsmock.ServiceLister)
		expErr  bool
		expResp tools.ListServicesToolOutput
	}{
		"It should map backend request and response.": {
			input: tools.ListServicesToolInput{Search: "check", Size: 100, Sort: string(backendapp.ServiceListSortModeAlertSeverityDesc), Cursor: "cursor-1"},
			mock: func(m *toolsmock.ServiceLister) {
				expReq := backendapp.ListServicesRequest{FilterSearchInput: "check", PageSize: 100, SortMode: backendapp.ServiceListSortModeAlertSeverityDesc, Cursor: "cursor-1"}
				m.On("ListServices", mock.Anything, expReq).Once().Return(&backendapp.ListServicesResponse{
					Services: []backendapp.ServiceAlerts{{
						Service: model.Service{ID: "checkout"},
						Stats: model.ServiceStats{
							TotalSLOs:                      4,
							SLOsCurrentlyBurningOverBudget: 1,
						},
						Alerts: []model.SLOAlerts{
							{FiringPage: &model.Alert{Name: "page"}},
							{FiringWarning: &model.Alert{Name: "warn"}},
						},
					}},
					PaginationCursors: backendapp.PaginationCursors{
						NextCursor: "next-1",
						HasNext:    true,
					},
				}, nil)
			},
			expResp: tools.ListServicesToolOutput{
				Services: []tools.ListServicesToolOutputItem{{
					ID:                             "checkout",
					TotalSLOs:                      4,
					SLOsCurrentlyBurningOverBudget: 1,
					TotalAlertsFiring:              2,
					HasWarning:                     true,
					HasCritical:                    true,
				}},
				Pagination: tools.ListSLOsToolOutputPagination{
					NextCursor: "next-1",
					HasNext:    true,
				},
			},
		},
		"It should default page size to 100.": {
			mock: func(m *toolsmock.ServiceLister) {
				m.On("ListServices", mock.Anything, mock.MatchedBy(func(req backendapp.ListServicesRequest) bool { return req.PageSize == 100 })).Once().Return(&backendapp.ListServicesResponse{}, nil)
			},
			expResp: tools.ListServicesToolOutput{Services: []tools.ListServicesToolOutputItem{}, Pagination: tools.ListSLOsToolOutputPagination{}},
		},
		"Having a backend error should fail.": {
			mock: func(m *toolsmock.ServiceLister) {
				m.On("ListServices", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("something wrong"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := toolsmock.NewServiceLister(t)
			if test.mock != nil {
				test.mock(m)
			}

			tool, handler := tools.NewListServicesTool(m, log.Noop)
			require.NotNil(t, tool)
			assert.Equal(t, "list_services", tool.Name)
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
