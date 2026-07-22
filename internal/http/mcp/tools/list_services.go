package tools

import (
	"context"
	"fmt"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	backendapp "github.com/slok/sloth/internal/http/backend/app"
	"github.com/slok/sloth/internal/log"
)

type ServiceLister interface {
	ListServices(ctx context.Context, req backendapp.ListServicesRequest) (*backendapp.ListServicesResponse, error)
}

func NewListServicesTool(app ServiceLister, logger log.Logger) (*sdkmcp.Tool, sdkmcp.ToolHandlerFor[ListServicesToolInput, ListServicesToolOutput]) {
	if logger == nil {
		logger = log.Noop
	}

	return &sdkmcp.Tool{
		Name:        "list_services",
		Description: "List services with fuzzy search, sorting, and pagination. Results include service IDs, total SLO counts, SLOs currently burning over budget, and the total number of firing alerts for each service. Search matches service IDs and defaults to 100 results per page.",
		Annotations: &sdkmcp.ToolAnnotations{ReadOnlyHint: true},
	}, newListServicesToolHandler(app, logger)
}

type ListServicesToolInput struct {
	Search string `json:"search,omitempty" jsonschema:"Optional fuzzy, case-insensitive search against service IDs. Spaces are ignored"`
	Size   int    `json:"size,omitempty" jsonschema:"Optional pagination size override, defaults to 100 and is capped at 100"`
	Sort   string `json:"sort,omitempty" jsonschema:"Sort mode: service-name-asc, service-name-desc, alert-severity-asc, alert-severity-desc"`
	Cursor string `json:"cursor,omitempty" jsonschema:"Pagination cursor returned by a previous response"`
}

type ListServicesToolOutput struct {
	Services   []ListServicesToolOutputItem `json:"services" jsonschema:"the services that matched the request"`
	Pagination ListSLOsToolOutputPagination `json:"pagination" jsonschema:"pagination cursors for requesting the next or previous page"`
}

type ListServicesToolOutputItem struct {
	ID                             string `json:"id" jsonschema:"the service ID"`
	TotalSLOs                      int    `json:"total_slos" jsonschema:"the total number of SLOs for the service"`
	SLOsCurrentlyBurningOverBudget int    `json:"slos_currently_burning_over_budget" jsonschema:"the number of SLOs currently burning over budget"`
	TotalAlertsFiring              int    `json:"total_alerts_firing" jsonschema:"the total number of firing alerts across the service SLOs"`
	HasWarning                     bool   `json:"has_warning" jsonschema:"whether any warning alert is firing for the service"`
	HasCritical                    bool   `json:"has_critical" jsonschema:"whether any page alert is firing for the service"`
}

func newListServicesToolHandler(app ServiceLister, logger log.Logger) sdkmcp.ToolHandlerFor[ListServicesToolInput, ListServicesToolOutput] {
	return func(ctx context.Context, _ *sdkmcp.CallToolRequest, input ListServicesToolInput) (*sdkmcp.CallToolResult, ListServicesToolOutput, error) {
		logger.WithValues(log.Kv{"input": fmt.Sprintf("%+v", input)}).Debugf("MCP tool called")

		pageSize := input.Size
		if pageSize <= 0 {
			pageSize = 100
		}

		resp, err := app.ListServices(ctx, backendapp.ListServicesRequest{
			FilterSearchInput: input.Search,
			PageSize:          pageSize,
			SortMode:          backendapp.ServiceListSortMode(input.Sort),
			Cursor:            input.Cursor,
		})
		if err != nil {
			return nil, ListServicesToolOutput{}, err
		}

		output := ListServicesToolOutput{
			Services: make([]ListServicesToolOutputItem, 0, len(resp.Services)),
			Pagination: ListSLOsToolOutputPagination{
				NextCursor:  resp.PaginationCursors.NextCursor,
				PrevCursor:  resp.PaginationCursors.PrevCursor,
				HasNext:     resp.PaginationCursors.HasNext,
				HasPrevious: resp.PaginationCursors.HasPrevious,
			},
		}

		for _, svc := range resp.Services {
			output.Services = append(output.Services, mapServiceToToolOutputItem(svc))
		}

		return nil, output, nil
	}
}

func mapServiceToToolOutputItem(svc backendapp.ServiceAlerts) ListServicesToolOutputItem {
	hasCritical := false
	hasWarning := false
	totalAlertsFiring := 0
	for _, sloAlert := range svc.Alerts {
		if sloAlert.FiringPage != nil {
			hasCritical = true
			totalAlertsFiring++
		}
		if sloAlert.FiringWarning != nil {
			hasWarning = true
			totalAlertsFiring++
		}
	}

	return ListServicesToolOutputItem{
		ID:                             svc.Service.ID,
		TotalSLOs:                      svc.Stats.TotalSLOs,
		SLOsCurrentlyBurningOverBudget: svc.Stats.SLOsCurrentlyBurningOverBudget,
		TotalAlertsFiring:              totalAlertsFiring,
		HasWarning:                     hasWarning,
		HasCritical:                    hasCritical,
	}
}
