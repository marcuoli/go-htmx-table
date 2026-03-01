# go-htmx-table

A reusable [templ](https://templ.guide)-based table component system for Go web applications using **HTMX + Alpine.js + Tailwind CSS**.

## Features

- **Server-side pagination**: First/Prev/Next/Last with page size selector
- **Server-side sorting**: clickable column headers with direction toggle
- **Search & filters**: debounced search input, dropdown filters
- **CRUD action buttons**: edit, delete, view, copy with SVG icons
- **Status badges**: auto-colored from status strings (active → green, error → red, etc.)
- **Dark mode**: Tailwind `dark:` variants throughout
- **Query param helpers**: `queryutil` sub-package for safe parsing of sort/page/filter params

## Installation

```bash
go get github.com/marcuoli/go-htmx-table@latest
```

## Quick Start

### Handler (Go)

```go
package handler

import (
    "net/http"

    table "github.com/marcuoli/go-htmx-table"
    "github.com/marcuoli/go-htmx-table/queryutil"
)

func CustomerList(w http.ResponseWriter, r *http.Request) {
    // Parse query params safely
    sortBy := queryutil.ResolveSortField(r.URL.Query().Get("sort_by"), []string{"name", "email"}, "name")
    sortDir := queryutil.NormalizeSortDirection(r.URL.Query().Get("sort_dir"), "asc")
    page := queryutil.ParsePage(r)
    pageSize := queryutil.ParsePageSize(r, 10, 100)

    // Fetch data from your API/DB...
    customers, totalItems := fetchCustomers(sortBy, sortDir, page, pageSize)
    totalPages := queryutil.TotalPages(totalItems, pageSize)

    // Build props and render
    props := table.DataTableProps{
        SortBy: sortBy, SortDir: sortDir,
        BaseURL: "/customer", Page: page, PageSize: pageSize,
        TotalPages: totalPages, TotalItems: totalItems,
    }
    CustomerTable(props, customers).Render(r.Context(), w)
}
```

### Template (templ)

```go
package handler

import table "github.com/marcuoli/go-htmx-table"

templ CustomerTable(props table.DataTableProps, customers []Customer) {
    // Filter bar
    @table.FilterCard("/customer") {
        @table.SearchInput("q", "Search customers...", "")
    }

    // Table
    @table.DataTable(props) {
        @table.DataTableHead() {
            @table.SortableCol(table.SortHeaderProps{
                Field: "name", Label: "Name",
                SortBy: props.SortBy, SortDir: props.SortDir,
                BaseURL: props.BaseURL,
            })
            @table.Col("Email")
            @table.ColRight("Actions")
        }
        @table.DataTableBody(table.BodyProps{ID: "customer-tbody"}) {
            for _, c := range customers {
                @table.Row() {
                    @table.Cell() { {c.Name} }
                    @table.CellMuted() { {c.Email} }
                    @table.CellActions() {
                        @table.ActionButton(table.ActionButtonProps{Action: "edit", URL: "/customer/" + c.ID + "/edit"})
                        @table.ActionButton(table.ActionButtonProps{Action: "delete", URL: "/customer/" + c.ID})
                    }
                }
            }
        }
    }

    // Pagination
    @table.ServerPagination(table.PaginationProps{
        Page: props.Page, TotalPages: props.TotalPages,
        TotalItems: props.TotalItems, PageSize: props.PageSize,
        BaseURL: props.BaseURL,
    })
}
```

## Components

### Table

| Component | Description |
|---|---|
| `DataTable(props)` | Table wrapper with `overflow-x-auto` |
| `DataTableHead()` | `<thead>` row |
| `DataTableBody(props)` | `<tbody>` with optional ID |
| `Row()` | Table row with hover effect |
| `Cell()` | Standard cell |
| `CellMuted()` | Secondary/muted cell |
| `CellActions()` | Right-aligned actions cell |
| `Col(label)` | Non-sortable column header |
| `ColRight(label)` | Right-aligned column header |
| `SortableCol(props)` | Clickable sortable column header |
| `EmptyState(message)` | Full-width empty message row |

### Pagination

| Component | Description |
|---|---|
| `ServerPagination(props)` | Full pagination bar with First/Prev/Next/Last + page size selector |

### Filters

| Component | Description |
|---|---|
| `FilterCard(action)` | Filter form wrapper with auto-submit |
| `SearchInput(name, placeholder, value)` | Debounced search field |
| `FilterSelect(name, label, selected, options)` | Dropdown filter |

### Actions

| Component | Description |
|---|---|
| `ActionButton(props)` | Icon button — edit (blue), delete (red), view (gray), copy (green) |
| `NewButton(props)` | "Add New" button with plus icon |
| `ActionButtonsGroup()` | Horizontal wrapper for action buttons |

### Badges

| Component | Description |
|---|---|
| `StatusBadge(label, variant)` | Colored badge with dot indicator |
| `StatusBadgeFromString(status)` | Auto-maps status text to color variant |

Badge variants: `BadgeSuccess`, `BadgeWarning`, `BadgeDanger`, `BadgeInfo`, `BadgeNeutral`

Auto-mapping: active/completed/done → success, pending/review → warning, error/failed → danger, draft/new → info

## Types

```go
type DataTableProps struct {
    ID, SortBy, SortDir, BaseURL string
    Page, PageSize, TotalPages, TotalItems int
}

type PaginationProps struct {
    Page, TotalPages, TotalItems, PageSize int
    BaseURL string
    PageSizes []int
}

type SortHeaderProps struct {
    Field, Label, SortBy, SortDir, BaseURL string
}

type ActionButtonProps struct {
    Action, URL, Confirm string
}

type FilterOption struct {
    Value, Label string
}
```

## Query Param Helpers (`queryutil`)

| Function | Description |
|---|---|
| `ResolveSortField(requested, allowed, default)` | Validates sort field against allow-list |
| `NormalizeSortDirection(dir, default)` | Ensures "asc" or "desc" |
| `ParsePage(r)` | Extracts page number (min 1) |
| `ParsePageSize(r, default, max)` | Extracts page size with clamping |
| `BuildOrderClause(field, dir)` | Returns `"field ASC/DESC"` |
| `TotalPages(total, pageSize)` | Calculates page count |
| `Offset(page, pageSize)` | Calculates SQL offset |

## URL Helpers (root package)

| Function | Description |
|---|---|
| `SortURL(baseURL, field, currentSort, currentDir)` | Builds URL that toggles sort direction |
| `PageURL(baseURL, page)` | Builds URL for a specific page |
| `PageSizeURL(baseURL, size)` | Builds URL for page size change (resets page) |
| `SortIcon(field, currentSort, currentDir)` | Returns ▲, ▼, or ↕ |
| `PaginationInfo(page, pageSize, totalItems)` | Returns "Showing 1–10 of 42 items" |

## Requirements

- Go 1.26+
- [templ](https://templ.guide) v0.3+
- Tailwind CSS 4
- HTMX 2.x
- Alpine.js (for filter auto-submit)

## License

MIT
