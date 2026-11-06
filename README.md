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
go get github.com/marcuoli/go-htmx-table@v0.2.1+20261105.19
```

Use the complete `vCORE+YYYYMMDD.PR` tag from the release. Go may record a
pseudo-version for that exact revision; `@latest` may select an older conventional
tag. Verify the resolved commit against the release rather than shortening its identity.

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

- Go 1.27.3+
- [templ](https://templ.guide) v0.3.1100 (matching generated code)
- Tailwind CSS 4
- HTMX 2.x
- Alpine.js (for filter auto-submit)

## Development

The templ generator and runtime are pinned together in go.mod:

```sh
go tool templ generate
go test ./...
go vet ./...
go mod verify
```

The new generated code requires templ v0.3.1100; older runtimes lack its
attribute-rendering APIs. All dependencies use official upstream sources.
The parse module retains v0.0.0-20250122154542-74294addb73e because the newer
v0.0.0-20261027111828-db2689052003 adds Parser.Match, which templ v0.3.1100's
utf8RuneParser does not implement. Recheck this upstream compatibility exception
on the next dependency review; see issue #12. No dependency source is patched.

### Module version planning

Use [go-versioning](https://github.com/marcuoli/go-versioning) release
`0.3.1+20261106.13` (Go coordinate `v0.3.2-0.20261106100520-4c3c478eb827`) to plan
full release identities:

```sh
go run ./scripts/module-version -previous v0.2.0 -bump patch -check v0.2.1+20261105.14 -commit-time 2026-11-05T12:00:00Z -pr 14
```

This illustrative plan emits release tag `v0.2.1+20261105.14` and full version text
`0.2.1+20261105.14`, matching OraScope's version/date/PR format. Supply the
actual target commit timestamp and merged PR number for a real build.
`go-versioning.Next` selects the date in `America/Sao_Paulo`; the planner
never invents missing commit metadata or uses the execution date.

The command uses `versioning.ParseCore` and `versioning.Next`, emits a JSON plan,
and accepts a full previous release tag or a legacy conventional module tag.
It rejects snapshot predecessors and mismatched expected tags. Choose
`inherit`, `major`, `minor`, or `patch` according to the change. It does not create tags
or publish releases. No bare-core alias is created. The JSON
`version` field provides the full table build text for release tooling. It is
not a lookup of the consuming executable's version. Validate releases with `GOWORK=off`
and the published dependency, without local replacements.

### Explicit tool builds and overall build numbers

The library never allocates a counter during imports, consumer compilation,
startup, tests or planning queries. The explicit project-owned tool build is:

```powershell
./scripts/build-module-version.ps1 -StorageRoot /absolute/project-storage -Store /absolute/live-authority -MainRepository /absolute/permanent-main
```

Use absolute external paths and PowerShell 7. One fresh request is persisted per
invocation before allocation; a failed compile consumes its reservation. The
pinned shared CLI validates the main VERSION location/ownership before any
reservation, injects metadata and owns all parsing, locking and journal
updates. `module-version -build-info` reads that metadata without Git/store access.
The output path and allocation/compiler logs are retained under external storage.
An ordinary `go build` bypasses this project workflow and has no known counter.

The wrapper builds a source snapshot by default. For a release artifact, first
verify the merged PR and its full tag, then pass `-ReleaseVersion FULL_VERSION`
to that same wrapper. The pinned CLI verifies release cleanliness/tag identity
and injects its shared linker flags into the one logical compilation. Verify
`-build-info` against the allocation. Preserve
the actual build time, counter ID and reservation; never relabel a snapshot
binary as a release or reuse an earlier compile's request ID.

Existing authorities use the official location-only correction after all builders
are upgraded and primary-main ignore rules are reviewed and integrated:

```sh
go tool go-versioning counter-status --repo /absolute/source --main-repo /absolute/permanent-main --store /absolute/live-authority --app go-htmx-table --json
go tool go-versioning counter-check --repo /absolute/source --main-repo /absolute/permanent-main --store /absolute/live-authority --app go-htmx-table --json
# Only the expected legacy-layout diagnostic proceeds to relocation:
go tool go-versioning counter-relocate --repo /absolute/source --main-repo /absolute/permanent-main --store /absolute/live-authority --app go-htmx-table --json
go tool go-versioning counter-check --repo /absolute/source --main-repo /absolute/permanent-main --store /absolute/live-authority --app go-htmx-table --json
```

Relocation creates missing root files at the existing number without allocating,
reseeding or changing identity/receipts. Other corruption/ownership/I/O failures
require investigation, never generic repair. A genuinely new authority uses the
official explicit-main initialization API only after the legacy-sequence inquiry;
the one-time complete canonical commit count is a fallback when no prior method
exists. Existing table state must never be initialized again.

`VERSION` is one decimal value plus LF at permanent primary main, beside
`VERSION.authority.json`. Both are untracked and ignored by reviewed root rules;
full release identity stays separate. Journals/locks/receipts remain in the same
external authority. Retired store VERSION evidence is preserved by the shared
migration, not used as an active counter. Never hand-edit, copy, reset or clone
these files or create per-worktree counters. Tests of allocation/migration require
a standalone isolated Git main fixture, never a linked worktree of live source.

All local/WSL/CI builders must reach the same actual main files and journal on the
allocation host and pass shared validation. No WSL/CI allocation entry point is
configured here; qualify that access before adding one. A runner clone, cloud-sync
copy or guessed mount alias cannot stand in for the real main. Old shared writers
refuse the relocated schema. Runtime/build-info/planning remain read-only.
See [the released location contract](https://github.com/marcuoli/go-versioning/blob/v0.3.1%2B20261106.13/docs/counter-location.md).

## License

MIT
