package tools_test

import (
	"context"
	"testing"

	"github.com/slok/sloth/internal/http/mcp/tools"
	"github.com/slok/sloth/internal/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewContextTool(t *testing.T) {
	tests := map[string]struct {
		expErr  bool
		expResp tools.ContextToolOutput
	}{
		"Tool should expose metadata and static context payload.": {
			expResp: tools.ContextToolOutput{
				Version:     "dev",
				Description: "Sloth is a Prometheus SLO framework that helps teams define service level objectives and creates a uniform, standardized layer of low-level Prometheus rules to implement SLOs, including the recording and alerting rules required to measure them.",
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			tool, handler := tools.NewContextTool(log.Noop)

			require.NotNil(t, tool)
			assert.Equal(t, "context", tool.Name)
			assert.Equal(t, "Get context about Sloth and its SLO framework. Returns the running Sloth version and a description of what Sloth does.", tool.Description)

			result, gotResp, err := handler(context.Background(), nil, tools.ContextToolInput{})
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
