package executions

import (
	"bytes"
	"fmt"
	"strings"
	"testing"

	"github.com/eduardolat/pgbackweb/internal/util/paginateutil"
	"github.com/google/uuid"
	nodx "github.com/nodxdev/nodxgo"
	"github.com/stretchr/testify/assert"
)

// defaultPaginationWindowSize is assumed to be 2 based on the component's implementation.
// This mirrors the constant in the original package.
const testDefaultPaginationWindowSize = 2

func renderNodxNode(t *testing.T, node nodx.Node) string {
	if node == nil {
		return ""
	}
	buf := bytes.Buffer{}
	err := node.Render(&buf)
	assert.NoError(t, err)
	return buf.String()
}

func TestPaginationComponent(t *testing.T) {
	baseQueryData := listExecsQueryData{}

	t.Run("No Pagination (Single Page / No Results)", func(t *testing.T) {
		tests := []struct {
			name         string
			pagination   paginateutil.PaginateResponse
			expectedHTML string
		}{
			{
				name:         "TotalPages = 0",
				pagination:   paginateutil.PaginateResponse{TotalPages: 0},
				expectedHTML: "",
			},
			{
				name:         "TotalPages = 1",
				pagination:   paginateutil.PaginateResponse{TotalPages: 1},
				expectedHTML: "",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := paginationComponent(baseQueryData, tt.pagination)
				htmlOutput := renderNodxNode(t, node)
				assert.Equal(t, tt.expectedHTML, htmlOutput)
			})
		}
	})

	t.Run("Few Pages", func(t *testing.T) {
		tests := []struct {
			name       string
			pagination paginateutil.PaginateResponse
			expected   []string // Substrings to check for
		}{
			{
				name:       "TotalPages = 3, CurrentPage = 1",
				pagination: paginateutil.PaginateResponse{TotalPages: 3, CurrentPage: 1},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn btn-active btn-disabled" disabled="true">First</button>`,
					`aria-label="Go to page 2" class="join-item btn" hx-get="/dashboard/executions/list?page=2">2</button>`,
					`aria-label="Go to last page, page 3" class="join-item btn" hx-get="/dashboard/executions/list?page=3">Last</button>`,
				},
			},
			{
				name:       "TotalPages = 3, CurrentPage = 2",
				pagination: paginateutil.PaginateResponse{TotalPages: 3, CurrentPage: 2},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn" hx-get="/dashboard/executions/list?page=1">First</button>`,
					`aria-label="Go to page 2" class="join-item btn btn-active btn-disabled" disabled="true">2</button>`,
					`aria-label="Go to last page, page 3" class="join-item btn" hx-get="/dashboard/executions/list?page=3">Last</button>`,
				},
			},
			{
				name:       "TotalPages = 3, CurrentPage = 3",
				pagination: paginateutil.PaginateResponse{TotalPages: 3, CurrentPage: 3},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn" hx-get="/dashboard/executions/list?page=1">First</button>`,
					`aria-label="Go to page 2" class="join-item btn" hx-get="/dashboard/executions/list?page=2">2</button>`,
					`aria-label="Go to last page, page 3" class="join-item btn btn-active btn-disabled" disabled="true">Last</button>`,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := paginationComponent(baseQueryData, tt.pagination)
				htmlOutput := renderNodxNode(t, node)
				assert.Contains(t, htmlOutput, `aria-label="Pagination navigation"`)
				for _, sub := range tt.expected {
					assert.Contains(t, htmlOutput, sub)
				}
			})
		}
	})

	t.Run("Many Pages (Windowing and Ellipses)", func(t *testing.T) {
		// defaultPaginationWindowSize is 2
		tests := []struct {
			name       string
			pagination paginateutil.PaginateResponse
			expected   []string
		}{
			{
				name:       "TotalPages = 20, CurrentPage = 1",
				pagination: paginateutil.PaginateResponse{TotalPages: 20, CurrentPage: 1},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn btn-active btn-disabled" disabled="true">First</button>`,
					`aria-label="Go to page 2" class="join-item btn" hx-get="/dashboard/executions/list?page=2">2</button>`,
					`aria-label="Go to page 3" class="join-item btn" hx-get="/dashboard/executions/list?page=3">3</button>`,
					`aria-label="Skipped pages" class="join-item btn btn-disabled" disabled="true">...</button>`,
					`aria-label="Go to last page, page 20" class="join-item btn" hx-get="/dashboard/executions/list?page=20">Last</button>`,
				},
			},
			{
				name:       "TotalPages = 20, CurrentPage = 10",
				pagination: paginateutil.PaginateResponse{TotalPages: 20, CurrentPage: 10},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn" hx-get="/dashboard/executions/list?page=1">First</button>`,
					`aria-label="Skipped pages" class="join-item btn btn-disabled" disabled="true">...</button>`, // Ellipsis before 8
					`aria-label="Go to page 8" class="join-item btn" hx-get="/dashboard/executions/list?page=8">8</button>`,
					`aria-label="Go to page 9" class="join-item btn" hx-get="/dashboard/executions/list?page=9">9</button>`,
					`aria-label="Go to page 10" class="join-item btn btn-active btn-disabled" disabled="true">10</button>`,
					`aria-label="Go to page 11" class="join-item btn" hx-get="/dashboard/executions/list?page=11">11</button>`,
					`aria-label="Go to page 12" class="join-item btn" hx-get="/dashboard/executions/list?page=12">12</button>`,
					`aria-label="Skipped pages" class="join-item btn btn-disabled" disabled="true">...</button>`, // Ellipsis after 12
					`aria-label="Go to last page, page 20" class="join-item btn" hx-get="/dashboard/executions/list?page=20">Last</button>`,
				},
			},
			{
				name:       "TotalPages = 20, CurrentPage = 20",
				pagination: paginateutil.PaginateResponse{TotalPages: 20, CurrentPage: 20},
				expected: []string{
					`aria-label="Go to first page" class="join-item btn" hx-get="/dashboard/executions/list?page=1">First</button>`,
					`aria-label="Skipped pages" class="join-item btn btn-disabled" disabled="true">...</button>`, // Ellipsis before 18
					`aria-label="Go to page 18" class="join-item btn" hx-get="/dashboard/executions/list?page=18">18</button>`,
					`aria-label="Go to page 19" class="join-item btn" hx-get="/dashboard/executions/list?page=19">19</button>`,
					`aria-label="Go to last page, page 20" class="join-item btn btn-active btn-disabled" disabled="true">Last</button>`,
				},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				node := paginationComponent(baseQueryData, tt.pagination)
				htmlOutput := renderNodxNode(t, node)
				assert.Contains(t, htmlOutput, `aria-label="Pagination navigation"`)
				// Check order and presence of all expected elements
				currentIndex := 0
				for _, sub := range tt.expected {
					index := strings.Index(htmlOutput[currentIndex:], sub)
					assert.True(t, index >= 0, "Substring not found or out of order: %s in %s", sub, htmlOutput[currentIndex:])
					if index >= 0 {
						currentIndex += index + len(sub)
					}
				}
			})
		}
	})

	t.Run("Query Parameter Preservation", func(t *testing.T) {
		backupID := uuid.New()
		queryDataWithBackup := listExecsQueryData{
			Backup: backupID,
		}
		pagination := paginateutil.PaginateResponse{TotalPages: 5, CurrentPage: 2}
		node := paginationComponent(queryDataWithBackup, pagination)
		htmlOutput := renderNodxNode(t, node)

		expectedQuerySuffix := fmt.Sprintf("&backup=%s", backupID.String())
		
		assert.Contains(t, htmlOutput, `aria-label="Pagination navigation"`)
		assert.Contains(t, htmlOutput, fmt.Sprintf(`hx-get="/dashboard/executions/list?page=1%s"`, expectedQuerySuffix)) // First button
		assert.Contains(t, htmlOutput, fmt.Sprintf(`hx-get="/dashboard/executions/list?page=3%s"`, expectedQuerySuffix)) // Page 3 (example)
		assert.Contains(t, htmlOutput, fmt.Sprintf(`hx-get="/dashboard/executions/list?page=5%s"`, expectedQuerySuffix)) // Last button
	})

	t.Run("ARIA Attributes and Active/Disabled States", func(t *testing.T) {
		// TotalPages = 5, CurrentPage = 3. WindowSize = 2.
		// Expected: [First/1] [2] [3 (active)] [4] [Last/5]
		pagination := paginateutil.PaginateResponse{TotalPages: 5, CurrentPage: 3}
		node := paginationComponent(baseQueryData, pagination)
		htmlOutput := renderNodxNode(t, node)

		// Container
		assert.Contains(t, htmlOutput, `<div class="join" aria-label="Pagination navigation">`)

		// First button (Page 1)
		assert.Contains(t, htmlOutput, `aria-label="Go to first page" class="join-item btn" hx-get="/dashboard/executions/list?page=1">First</button>`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to first page" class="join-item btn" disabled="true"`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to first page" class="join-item btn" aria-current="page"`)
		
		// Page 2 button
		assert.Contains(t, htmlOutput, `aria-label="Go to page 2" class="join-item btn" hx-get="/dashboard/executions/list?page=2">2</button>`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to page 2" class="join-item btn" disabled="true"`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to page 2" class="join-item btn" aria-current="page"`)

		// Page 3 button (Current)
		assert.Contains(t, htmlOutput, `aria-label="Go to page 3" class="join-item btn btn-active btn-disabled" aria-current="page" disabled="true">3</button>`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to page 3" class="join-item btn btn-active btn-disabled" hx-get`) // No hx-get for current

		// Page 4 button
		assert.Contains(t, htmlOutput, `aria-label="Go to page 4" class="join-item btn" hx-get="/dashboard/executions/list?page=4">4</button>`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to page 4" class="join-item btn" disabled="true"`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to page 4" class="join-item btn" aria-current="page"`)

		// Last button (Page 5)
		assert.Contains(t, htmlOutput, `aria-label="Go to last page, page 5" class="join-item btn" hx-get="/dashboard/executions/list?page=5">Last</button>`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to last page, page 5" class="join-item btn" disabled="true"`)
		assert.NotContains(t, htmlOutput, `aria-label="Go to last page, page 5" class="join-item btn" aria-current="page"`)
		
		// No ellipsis should be present in this specific case (TotalPages = 5, CurrentPage = 3, windowSize = 2)
		// Window: [C-2, C-1, C, C+1, C+2] -> [1, 2, 3, 4, 5]
		// leftWindowBound = 3-2 = 1. rightWindowBound = 3+2 = 5
		// Left ellipsis: leftWindowBound > 2 (1 > 2) is false.
		// Right ellipsis: rightWindowBound < totalPages - 1 (5 < 4) is false.
		assert.NotContains(t, htmlOutput, `aria-label="Skipped pages"`)
	})

	t.Run("ARIA Attributes for Ellipsis", func(t *testing.T) {
		// TotalPages = 7, CurrentPage = 1. WindowSize = 2
		// Expected: [First/1 (active)] [2] [3] [...] [Last/7]
		pagination := paginateutil.PaginateResponse{TotalPages: 7, CurrentPage: 1}
		node := paginationComponent(baseQueryData, pagination)
		htmlOutput := renderNodxNode(t, node)

		assert.Contains(t, htmlOutput, `aria-label="Skipped pages" class="join-item btn btn-disabled" disabled="true">...</button>`)
	})
}
