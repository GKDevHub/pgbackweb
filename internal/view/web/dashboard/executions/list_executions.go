package executions

import (
	"fmt"
	"net/http"

	"github.com/eduardolat/pgbackweb/internal/database/dbgen"
	"github.com/eduardolat/pgbackweb/internal/service/executions"
	"github.com/eduardolat/pgbackweb/internal/util/echoutil"
	"github.com/eduardolat/pgbackweb/internal/util/paginateutil"
	"github.com/eduardolat/pgbackweb/internal/util/strutil"
	"github.com/eduardolat/pgbackweb/internal/util/timeutil"
	"github.com/eduardolat/pgbackweb/internal/validate"
	"github.com/eduardolat/pgbackweb/internal/view/web/component"
	"github.com/eduardolat/pgbackweb/internal/view/web/respondhtmx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	nodx "github.com/nodxdev/nodxgo"
	htmx "github.com/nodxdev/nodxgo-htmx"
)

type listExecsQueryData struct {
	Database    uuid.UUID `query:"database" validate:"omitempty,uuid"`
	Destination uuid.UUID `query:"destination" validate:"omitempty,uuid"`
	Backup      uuid.UUID `query:"backup" validate:"omitempty,uuid"`
	Page        int       `query:"page" validate:"required,min=1"`
}

func (h *handlers) listExecutionsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	var queryData listExecsQueryData
	if err := c.Bind(&queryData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}
	if err := validate.Struct(&queryData); err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	pagination, executions, err := h.servs.ExecutionsService.PaginateExecutions(
		ctx, executions.PaginateExecutionsParams{
			DatabaseFilter: uuid.NullUUID{
				UUID: queryData.Database, Valid: queryData.Database != uuid.Nil,
			},
			DestinationFilter: uuid.NullUUID{
				UUID: queryData.Destination, Valid: queryData.Destination != uuid.Nil,
			},
			BackupFilter: uuid.NullUUID{
				UUID: queryData.Backup, Valid: queryData.Backup != uuid.Nil,
			},
			Page:  queryData.Page,
			Limit: 20,
		},
	)
	if err != nil {
		return respondhtmx.ToastError(c, err.Error())
	}

	return echoutil.RenderNodx(
		c, http.StatusOK, listExecutions(queryData, pagination, executions),
	)
}

func listExecutions(
	queryData listExecsQueryData,
	pagination paginateutil.PaginateResponse,
	executions []dbgen.ExecutionsServicePaginateExecutionsRow,
) nodx.Node {
	if len(executions) < 1 {
		return component.EmptyResultsTr(component.EmptyResultsParams{
			Title:    "No executions found",
			Subtitle: "Wait for the first execution to appear here",
		})
	}

	trs := []nodx.Node{}
	for _, execution := range executions {
		trs = append(trs, nodx.Tr(
			nodx.Td(component.OptionsDropdown(
				showExecutionButton(execution),
				restoreExecutionButton(execution),
			)),
			nodx.Td(component.StatusBadge(execution.Status)),
			nodx.Td(component.SpanText(execution.BackupName)),
			nodx.Td(component.SpanText(execution.DatabaseName)),
			nodx.Td(component.PrettyDestinationName(
				execution.BackupIsLocal, execution.DestinationName,
			)),
			nodx.Td(component.SpanText(
				execution.StartedAt.Local().Format(timeutil.LayoutYYYYMMDDHHMMSSPretty),
			)),
			nodx.Td(
				nodx.If(
					execution.FinishedAt.Valid,
					component.SpanText(
						execution.FinishedAt.Time.Local().Format(timeutil.LayoutYYYYMMDDHHMMSSPretty),
					),
				),
			),
			nodx.Td(
				nodx.If(
					execution.FinishedAt.Valid,
					component.SpanText(
						execution.FinishedAt.Time.Sub(execution.StartedAt).String(),
					),
				),
			),
			nodx.Td(
				nodx.If(
					execution.FileSize.Valid,
					component.PrettyFileSize(execution.FileSize),
				),
			),
		))
	}

	tableRows := component.RenderableGroup(trs)
	paginationComp := paginationComponent(queryData, pagination)

	return nodx.Group(tableRows, paginationComp)
}

const defaultPaginationWindowSize = 2

// createPageButtonNode is a helper function to create a pagination button.
// text: Text content of the button (e.g., "1", "First", "Last").
// pageNum: The page number this button links to. For placeholders, this can be 0.
// isPlaceholder: True if this button is an ellipsis ("...").
// isCurrent: True if this button represents the current active page.
// totalPages: Total number of pages, used for ARIA label on "Last" button.
// queryData: Used to construct HxGet URL with existing filters.
func createPageButtonNode(text string, pageNum int, isPlaceholder bool, isCurrent bool, totalPages int, queryData listExecsQueryData) nodx.Node {
	btnClass := "join-item btn"
	if isCurrent && !isPlaceholder {
		btnClass += " btn-active"
	}

	isInteractive := !isPlaceholder && !isCurrent
	if !isInteractive {
		btnClass += " btn-disabled"
	}

	attrs := []nodx.Node{nodx.Class(btnClass)}

	// ARIA Label
	var ariaLabel string
	if text == "First" {
		ariaLabel = "Go to first page"
	} else if text == "Last" {
		ariaLabel = fmt.Sprintf("Go to last page, page %d", totalPages)
	} else if isPlaceholder {
		ariaLabel = "Skipped pages"
	} else {
		ariaLabel = fmt.Sprintf("Go to page %d", pageNum)
	}
	attrs = append(attrs, nodx.Attr("aria-label", ariaLabel))

	// ARIA Current
	if isCurrent && !isPlaceholder {
		attrs = append(attrs, nodx.Attr("aria-current", "page"))
	}

	// Standard HTML disabled attribute for non-interactive buttons
	if !isInteractive {
		attrs = append(attrs, nodx.Attr("disabled", "true"))
	}

	if isPlaceholder {
		attrs = append(attrs, nodx.Text("..."))
	} else {
		attrs = append(attrs, nodx.Text(text))
	}

	if isInteractive {
		attrs = append(attrs,
			htmx.HxGet(func() string {
				url := "/dashboard/executions/list"
				url = strutil.AddQueryParamToUrl(url, "page", fmt.Sprintf("%d", pageNum))
				if queryData.Database != uuid.Nil {
					url = strutil.AddQueryParamToUrl(url, "database", queryData.Database.String())
				}
				if queryData.Destination != uuid.Nil {
					url = strutil.AddQueryParamToUrl(url, "destination", queryData.Destination.String())
				}
				if queryData.Backup != uuid.Nil {
					url = strutil.AddQueryParamToUrl(url, "backup", queryData.Backup.String())
				}
				return url
			}()),
			htmx.HxTarget("tbody"),
			htmx.HxSwap("innerHTML"),
		)
	}
	return nodx.Button(attrs...)
}

func paginationComponent(
	queryData listExecsQueryData,
	pagination paginateutil.PaginateResponse,
) nodx.Node {
	if pagination.TotalPages <= 1 {
		return nil
	}

	buttons := []nodx.Node{}
	pagesShown := make(map[int]bool)

	// 1. "First" Button
	buttons = append(buttons, createPageButtonNode("First", 1, false, pagination.CurrentPage == 1, pagination.TotalPages, queryData))
	pagesShown[1] = true

	// 2. Window and Ellipses
	leftWindowBound := pagination.CurrentPage - defaultPaginationWindowSize
	rightWindowBound := pagination.CurrentPage + defaultPaginationWindowSize

	// Left Ellipsis
	// Show ellipsis if the window starts after page 2 (i.e., page 1, then ..., then window)
	if leftWindowBound > 2 {
		buttons = append(buttons, createPageButtonNode("...", 0, true, false, pagination.TotalPages, queryData))
	}

	// Window Pages
	for page := leftWindowBound; page <= rightWindowBound; page++ {
		// Only show valid pages that are not First or Last (those are handled separately)
		// and are within the overall page range.
		if page > 1 && page < pagination.TotalPages && page >= 1 {
			if !pagesShown[page] {
				buttons = append(buttons, createPageButtonNode(fmt.Sprintf("%d", page), page, false, page == pagination.CurrentPage, pagination.TotalPages, queryData))
				pagesShown[page] = true
			}
		}
	}

	// Right Ellipsis
	// Show ellipsis if the window ends before (TotalPages - 1)
	if rightWindowBound < pagination.TotalPages-1 {
		buttons = append(buttons, createPageButtonNode("...", 0, true, false, pagination.TotalPages, queryData))
	}

	// 3. "Last" Button
	// Add if TotalPages > 1. The pagesShown check is not needed because the window loop
	// (page > 1 && page < pagination.TotalPages) ensures "Last" is never part of it.
	if pagination.TotalPages > 1 {
		buttons = append(buttons, createPageButtonNode("Last", pagination.TotalPages, false, pagination.CurrentPage == pagination.TotalPages, pagination.TotalPages, queryData))
	}

	return nodx.Div(
		nodx.Class("join"),
		nodx.Attr("aria-label", "Pagination navigation"),
		nodx.Group(buttons...),
	)
}
