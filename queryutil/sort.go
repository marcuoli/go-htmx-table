// Package queryutil provides server-side helpers for parsing table query
// parameters (sorting, pagination, filtering) from HTTP requests.
package queryutil

import (
	"net/http"
	"strconv"
)

// ResolveSortField validates the requested sort field against an allow-list.
// If the requested field is not in the allowed list, it returns defaultField.
func ResolveSortField(requested string, allowed []string, defaultField string) string {
	if requested == "" {
		return defaultField
	}
	for _, f := range allowed {
		if f == requested {
			return requested
		}
	}
	return defaultField
}

// NormalizeSortDirection normalizes a sort direction string to "asc" or "desc".
func NormalizeSortDirection(dir, defaultDir string) string {
	switch dir {
	case "asc", "desc":
		return dir
	default:
		if defaultDir == "desc" {
			return "desc"
		}
		return "asc"
	}
}

// ParsePage extracts and validates the page number from query params.
// Returns at minimum 1.
func ParsePage(r *http.Request) int {
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		return 1
	}
	return page
}

// ParsePageSize extracts and validates the page size from query params.
// Falls back to defaultSize. Clamps between 1 and maxSize.
func ParsePageSize(r *http.Request, defaultSize, maxSize int) int {
	ps, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || ps < 1 {
		return defaultSize
	}
	if ps > maxSize {
		return maxSize
	}
	return ps
}

// BuildOrderClause builds a SQL-compatible ORDER BY clause string.
// Example: BuildOrderClause("name", "asc") returns "name ASC".
func BuildOrderClause(field, direction string) string {
	dir := "ASC"
	if direction == "desc" {
		dir = "DESC"
	}
	return field + " " + dir
}

// TotalPages calculates the number of pages for a given total and page size.
func TotalPages(totalItems, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	pages := totalItems / pageSize
	if totalItems%pageSize > 0 {
		pages++
	}
	return pages
}

// Offset calculates the SQL OFFSET for a given page and page size.
func Offset(page, pageSize int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * pageSize
}