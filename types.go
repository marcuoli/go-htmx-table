// Package table provides reusable templ-based table, pagination, sorting,
// and filter components for Go web applications using HTMX + Alpine.js + Tailwind CSS.
package table

import (
	"fmt"
	"net/url"
	"strconv"
)

// ---------------------------------------------------------------------------
// Core props
// ---------------------------------------------------------------------------

// DataTableProps configures the server-side table.
type DataTableProps struct {
	ID         string // HTML id for the <table> element (default: "data-table")
	SortBy     string // Current sort field
	SortDir    string // "asc" or "desc"
	BaseURL    string // Base URL for sort/page links
	Page       int    // Current page (1-based)
	PageSize   int    // Items per page
	TotalPages int    // Total number of pages
	TotalItems int    // Total records
}

// TableID returns the HTML id, defaulting to "data-table".
func (p DataTableProps) TableID() string {
	if p.ID != "" {
		return p.ID
	}
	return "data-table"
}

// PaginationProps for the pagination bar.
type PaginationProps struct {
	Page       int
	TotalPages int
	TotalItems int
	PageSize   int
	BaseURL    string
	PageSizes  []int  // e.g. [10, 20, 50, 100]
	HxTarget   string // HTMX target selector (defaults to "#content")
	HxPushURL  bool   // Whether to update browser URL on page change
}

// Target returns the HTMX target, defaulting to "#content".
func (p PaginationProps) Target() string {
	if p.HxTarget != "" {
		return p.HxTarget
	}
	return "#content"
}

// SortHeaderProps for sortable column headers.
type SortHeaderProps struct {
	Field     string // Field name for sorting
	Label     string // Display text
	SortBy    string // Currently active sort field
	SortDir   string // Current sort direction
	BaseURL   string // URL with preserved query params
	HxTarget  string // HTMX target selector (defaults to "#content")
	HxPushURL bool   // Whether to update browser URL on sort
}

// Target returns the HTMX target, defaulting to "#content".
func (p SortHeaderProps) Target() string {
	if p.HxTarget != "" {
		return p.HxTarget
	}
	return "#content"
}

// BodyProps configures the <tbody> element.
type BodyProps struct {
	ID string // HTML id for the <tbody> element
}

// FilterProps for the filter card.
type FilterProps struct {
	Action string // Form action URL
	Method string // HTTP method (default: GET)
}

// FilterOption represents a single <option> in a filter select.
type FilterOption struct {
	Value string
	Label string
}

// FilterCardProps configures the filter/search card above a table.
type FilterCardProps struct {
	Action    string // Form action URL (also hx-get target)
	HxTarget  string // HTMX target selector (defaults to "#content")
	HxPushURL bool   // Whether to update browser URL on filter change
}

// Target returns the HTMX target, defaulting to "#content".
func (p FilterCardProps) Target() string {
	if p.HxTarget != "" {
		return p.HxTarget
	}
	return "#content"
}

// ---------------------------------------------------------------------------
// Action / button props
// ---------------------------------------------------------------------------

// ActionButtonProps configures a CRUD action button (edit, delete, view, etc.).
type ActionButtonProps struct {
	Action  string // "edit", "delete", "view", "copy"
	URL     string // Target URL or hx-get/hx-delete target
	Confirm string // Optional confirmation message (for delete)
}

// NewButtonProps configures the "Add new" button above the table.
type NewButtonProps struct {
	Label string // Button text (default: "Add New")
	URL   string // hx-get target for the create form
}

// ---------------------------------------------------------------------------
// Single-call list API
// ---------------------------------------------------------------------------

// ListColumn defines one table column in the single-call List component.
type ListColumn struct {
	Field    string // Field name used for sorting
	Label    string // Header label
	Sortable bool   // Whether this column is sortable
	Align    string // "left" (default) or "right"
}

// ListClasses defines all CSS classes used by the single-call List component.
// The client application owns these values and passes them through ListProps.
type ListClasses struct {
	FilterCard      string
	FilterForm      string
	SearchWrap      string
	SearchInput     string
	Wrapper         string
	TableWrap       string
	Table           string
	Head            string
	HeaderCell      string
	HeaderCellRight string
	SortLink        string
	SortIcon        string
	Body            string
	EmptyCell       string
	EmptyWrap       string
	EmptyIcon       string
	EmptyText       string
	Pagination      string
	PaginationInfo  string
	PaginationNav   string
	PageIndicator   string
	PageSizeWrap    string
	PageSizeLabel   string
	PageSizeSelect  string
	PageBtn         string
	PageBtnActive   string
	PageBtnDisabled string
}

// ListProps configures the high-level List component.
type ListProps struct {
	TableID            string
	BodyID             string
	Columns            []ListColumn
	SortBy             string
	SortDir            string
	BaseURL            string
	Page               int
	PageSize           int
	TotalPages         int
	TotalItems         int
	PageSizes          []int
	Search             string
	SearchName         string
	SearchPlaceholder  string
	SearchEnabled      bool
	HasRows            bool
	EmptyMessage       string
	HxTarget           string
	HxPushURL          bool
	RefreshTrigger     string
	RefreshURL         string
	Classes            ListClasses
}

// Target returns the HTMX target, defaulting to "#content".
func (p ListProps) Target() string {
	if p.HxTarget != "" {
		return p.HxTarget
	}
	return "#content"
}

// EffectiveSearchName returns the input name for search query.
func (p ListProps) EffectiveSearchName() string {
	if p.SearchName != "" {
		return p.SearchName
	}
	return "search"
}

// EffectiveEmptyMessage returns the empty state message.
func (p ListProps) EffectiveEmptyMessage() string {
	if p.EmptyMessage != "" {
		return p.EmptyMessage
	}
	return "No items found"
}

// EffectiveTableID returns table id with fallback.
func (p ListProps) EffectiveTableID() string {
	if p.TableID != "" {
		return p.TableID
	}
	return "data-table"
}

// ---------------------------------------------------------------------------
// Badge props
// ---------------------------------------------------------------------------

// BadgeVariant controls the color scheme of a StatusBadge.
type BadgeVariant string

const (
	BadgeSuccess BadgeVariant = "success"
	BadgeWarning BadgeVariant = "warning"
	BadgeDanger  BadgeVariant = "danger"
	BadgeInfo    BadgeVariant = "info"
	BadgeNeutral BadgeVariant = "neutral"
)

// ---------------------------------------------------------------------------
// Defaults
// ---------------------------------------------------------------------------

// DefaultPageSizes returns the standard page size options.
func DefaultPageSizes() []int {
	return []int{10, 20, 50, 100}
}

// ---------------------------------------------------------------------------
// URL helpers
// ---------------------------------------------------------------------------

// SortURL builds a URL that toggles the sort direction for a given field.
// If the field is already the active sort, it flips the direction; otherwise
// it defaults to ascending.
func SortURL(baseURL, field, currentSort, currentDir string) string {
	dir := "asc"
	if field == currentSort && currentDir == "asc" {
		dir = "desc"
	}
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	q := u.Query()
	q.Set("sort_by", field)
	q.Set("sort_dir", dir)
	q.Del("page") // reset to page 1 on sort change
	u.RawQuery = q.Encode()
	return u.String()
}

// PageURL builds a URL for a specific page number.
func PageURL(baseURL string, page int) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	q := u.Query()
	q.Set("page", strconv.Itoa(page))
	u.RawQuery = q.Encode()
	return u.String()
}

// PageSizeURL builds a URL for changing the page size (resets to page 1).
func PageSizeURL(baseURL string, size int) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	q := u.Query()
	q.Set("page_size", strconv.Itoa(size))
	q.Del("page") // reset to page 1
	u.RawQuery = q.Encode()
	return u.String()
}

// SortIcon returns a Unicode arrow indicating the current sort direction
// for a field: ▲ (asc), ▼ (desc), or ↕ (not sorted by this field).
func SortIcon(field, currentSort, currentDir string) string {
	if field != currentSort {
		return "↕"
	}
	if currentDir == "desc" {
		return "▼"
	}
	return "▲"
}

// PaginationInfo returns a human-readable string like "Showing 1–10 of 42 items".
func PaginationInfo(page, pageSize, totalItems int) string {
	if totalItems == 0 {
		return "No items"
	}
	start := (page-1)*pageSize + 1
	end := page * pageSize
	if end > totalItems {
		end = totalItems
	}
	return fmt.Sprintf("Showing %d\u2013%d of %d items", start, end, totalItems)
}
