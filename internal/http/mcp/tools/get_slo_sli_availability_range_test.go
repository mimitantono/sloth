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

func TestNewGetSLOSLIAvailabilityRangeTool(t *testing.T) {
	ts1 := time.Date(2026, 5, 25, 10, 0, 0, 0, time.UTC)
	ts2 := time.Date(2026, 5, 25, 11, 0, 0, 0, time.UTC)

	tests := map[string]struct {
		input   tools.GetSLOSLIAvailabilityRangeToolInput
		mock    func(m *toolsmock.SLIAvailabilityRangeLister)
		expErr  bool
		expResp tools.GetSLOSLIAvailabilityRangeToolOutput
	}{
		"It should map the backend request and response.": {
			input: tools.GetSLOSLIAvailabilityRangeToolInput{SLOID: "slo-id", From: ts1.Format(time.RFC3339), To: ts2.Add(time.Hour).Format(time.RFC3339)},
			mock: func(m *toolsmock.SLIAvailabilityRangeLister) {
				expReq := backendapp.ListSLIAvailabilityRangeRequest{SLOID: "slo-id", From: ts1, To: ts2.Add(time.Hour)}
				m.On("ListSLIAvailabilityRange", mock.Anything, expReq).Once().Return(&backendapp.ListSLIAvailabilityRangeResponse{
					AvailabilityDataPoints: []model.DataPoint{{TS: ts1, Value: 99.9}, {TS: ts2, Missing: true}, {TS: ts2.Add(time.Hour), Value: 98.7}},
				}, nil)
			},
			expResp: tools.GetSLOSLIAvailabilityRangeToolOutput{StartTS: ts1.Format(time.RFC3339), Step: "1h0m0s", AvailabilitySeries: "99.9,x,98.7"},
		},
		"Invalid from time should fail.": {
			input:  tools.GetSLOSLIAvailabilityRangeToolInput{SLOID: "slo-id", From: "bad-time"},
			expErr: true,
		},
		"Invalid to time should fail.": {
			input:  tools.GetSLOSLIAvailabilityRangeToolInput{SLOID: "slo-id", From: ts1.Format(time.RFC3339), To: "bad-time"},
			expErr: true,
		},
		"Having a backend error should fail.": {
			input: tools.GetSLOSLIAvailabilityRangeToolInput{SLOID: "slo-id", From: ts1.Format(time.RFC3339)},
			mock: func(m *toolsmock.SLIAvailabilityRangeLister) {
				m.On("ListSLIAvailabilityRange", mock.Anything, mock.Anything).Once().Return(nil, fmt.Errorf("something wrong"))
			},
			expErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			m := toolsmock.NewSLIAvailabilityRangeLister(t)
			if test.mock != nil {
				test.mock(m)
			}

			tool, handler := tools.NewGetSLOSLIAvailabilityRangeTool(m, log.Noop)
			require.NotNil(t, tool)
			assert.Equal(t, "get_slo_sli_availability_range", tool.Name)
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
