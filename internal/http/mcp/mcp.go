package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/http/mcp/tools"
	"github.com/slok/sloth/internal/http/ui"
	"github.com/slok/sloth/internal/info"
	"github.com/slok/sloth/internal/log"
)

type ServiceApp interface {
	ListServices(ctx context.Context, req backendapp.ListServicesRequest) (*backendapp.ListServicesResponse, error)
	ListSLOs(ctx context.Context, req backendapp.ListSLOsRequest) (*backendapp.ListSLOsResponse, error)
	GetSLO(ctx context.Context, req backendapp.GetSLORequest) (*backendapp.GetSLOResponse, error)
	ListSLIAvailabilityRange(ctx context.Context, req backendapp.ListSLIAvailabilityRangeRequest) (*backendapp.ListSLIAvailabilityRangeResponse, error)
	ListBurnedBudgetRange(ctx context.Context, req backendapp.ListBurnedBudgetRangeRequest) (*backendapp.ListBurnedBudgetRangeResponse, error)
}

type Config struct {
	Logger     log.Logger
	ServiceApp ServiceApp
}

func (c *Config) defaults() error {
	if c.Logger == nil {
		c.Logger = log.Noop
	}
	c.Logger = c.Logger.WithValues(log.Kv{"component": "mcp"})
	if c.ServiceApp == nil {
		return fmt.Errorf("service app is required")
	}

	return nil
}

func New(cfg Config) (http.Handler, error) {
	err := cfg.defaults()
	if err != nil {
		return nil, err
	}

	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "sloth",
		Version: info.Version,
		Icons:   defaultIcons,
	}, nil)
	registeredTools := 0

	contextTool, contextToolHandler := tools.NewContextTool(cfg.Logger.WithValues(log.Kv{"tool": "context"}))
	registerTool(server, contextTool, contextToolHandler)
	registeredTools++
	listSLOsTool, listSLOsToolHandler := tools.NewListSLOsTool(cfg.ServiceApp, cfg.Logger.WithValues(log.Kv{"tool": "list_slos"}))
	registerTool(server, listSLOsTool, listSLOsToolHandler)
	registeredTools++
	listServicesTool, listServicesToolHandler := tools.NewListServicesTool(cfg.ServiceApp, cfg.Logger.WithValues(log.Kv{"tool": "list_services"}))
	registerTool(server, listServicesTool, listServicesToolHandler)
	registeredTools++
	getSLOTool, getSLOToolHandler := tools.NewGetSLOTool(cfg.ServiceApp, cfg.Logger.WithValues(log.Kv{"tool": "get_slo"}))
	registerTool(server, getSLOTool, getSLOToolHandler)
	registeredTools++
	getBurnedBudgetRangeTool, getBurnedBudgetRangeToolHandler := tools.NewGetSLOBurnedBudgetRangeTool(cfg.ServiceApp, cfg.Logger.WithValues(log.Kv{"tool": "get_slo_burned_budget_range"}))
	registerTool(server, getBurnedBudgetRangeTool, getBurnedBudgetRangeToolHandler)
	registeredTools++
	getSLIAvailabilityRangeTool, getSLIAvailabilityRangeToolHandler := tools.NewGetSLOSLIAvailabilityRangeTool(cfg.ServiceApp, cfg.Logger.WithValues(log.Kv{"tool": "get_slo_sli_availability_range"}))
	registerTool(server, getSLIAvailabilityRangeTool, getSLIAvailabilityRangeToolHandler)
	registeredTools++

	cfg.Logger.WithValues(log.Kv{"tools": registeredTools}).Infof("MCP request/response handler enabled")

	return sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
		return server
	}, &sdkmcp.StreamableHTTPOptions{
		Stateless:    true,
		JSONResponse: true,
	}), nil
}

func newSlothIcon(color string, theme sdkmcp.IconTheme) sdkmcp.Icon {
	svg := ui.SlothLogoSVG(color, 256)
	b64 := base64.StdEncoding.EncodeToString([]byte(svg))

	return sdkmcp.Icon{
		Source:   "data:image/svg+xml;base64," + b64,
		MIMEType: "image/svg+xml",
		Sizes:    []string{"any"},
		Theme:    theme,
	}
}

var defaultIcons = []sdkmcp.Icon{
	newSlothIcon("#111111", sdkmcp.IconThemeLight),
	newSlothIcon("#ffffff", sdkmcp.IconThemeDark),
}

func registerTool[In, Out any](server *sdkmcp.Server, tool *sdkmcp.Tool, handler sdkmcp.ToolHandlerFor[In, Out]) {
	if len(tool.Icons) == 0 {
		tool.Icons = defaultIcons
	}

	sdkmcp.AddTool(server, tool, handler)
}
