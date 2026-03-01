package queryutil

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestResolveSortField_Valid(t *testing.T) {
	got := ResolveSortField("email", []string{"name", "email", "created"}, "name")
	if got != "email" {
		t.Errorf("ResolveSortField() = %q; want %q", got, "email")
	}
}

func TestResolveSortField_Invalid(t *testing.T) {
	got := ResolveSortField("hacked", []string{"name", "email"}, "name")
	if got != "name" {
		t.Errorf("ResolveSortField() = %q; want %q", got, "name")
	}
}

func TestResolveSortField_Empty(t *testing.T) {
	got := ResolveSortField("", []string{"name"}, "name")
	if got != "name" {
		t.Errorf("ResolveSortField() = %q; want %q", got, "name")
	}
}

func TestNormalizeSortDirection(t *testing.T) {
	tests := []struct {
		dir, def, want string
	}{
		{"asc", "asc", "asc"},
		{"desc", "asc", "desc"},
		{"invalid", "asc", "asc"},
		{"invalid", "desc", "desc"},
		{"", "asc", "asc"},
	}
	for _, tc := range tests {
		got := NormalizeSortDirection(tc.dir, tc.def)
		if got != tc.want {
			t.Errorf("NormalizeSortDirection(%q, %q) = %q; want %q", tc.dir, tc.def, got, tc.want)
		}
	}
}

func makeRequest(url string) *http.Request {
	r := httptest.NewRequest("GET", url, nil)
	return r
}

func TestParsePage(t *testing.T) {
	tests := []struct {
		url  string
		want int
	}{
		{"/?page=3", 3},
		{"/?page=0", 1},
		{"/?page=-1", 1},
		{"/?page=abc", 1},
		{"/", 1},
	}
	for _, tc := range tests {
		got := ParsePage(makeRequest(tc.url))
		if got != tc.want {
			t.Errorf("ParsePage(%q) = %d; want %d", tc.url, got, tc.want)
		}
	}
}

func TestParsePageSize(t *testing.T) {
	tests := []struct {
		url      string
		def, max int
		want     int
	}{
		{"/?page_size=20", 10, 100, 20},
		{"/?page_size=200", 10, 100, 100}, // clamped
		{"/?page_size=0", 10, 100, 10},    // fallback
		{"/?page_size=abc", 10, 100, 10},  // fallback
		{"/", 10, 100, 10},                // fallback
	}
	for _, tc := range tests {
		got := ParsePageSize(makeRequest(tc.url), tc.def, tc.max)
		if got != tc.want {
			t.Errorf("ParsePageSize(%q, %d, %d) = %d; want %d", tc.url, tc.def, tc.max, got, tc.want)
		}
	}
}

func TestBuildOrderClause(t *testing.T) {
	if got := BuildOrderClause("name", "asc"); got != "name ASC" {
		t.Errorf("BuildOrderClause() = %q; want %q", got, "name ASC")
	}
	if got := BuildOrderClause("created", "desc"); got != "created DESC" {
		t.Errorf("BuildOrderClause() = %q; want %q", got, "created DESC")
	}
}

func TestTotalPages(t *testing.T) {
	tests := []struct {
		total, size, want int
	}{
		{42, 10, 5},
		{40, 10, 4},
		{0, 10, 0},
		{10, 0, 0}, // edge: zero page size
	}
	for _, tc := range tests {
		got := TotalPages(tc.total, tc.size)
		if got != tc.want {
			t.Errorf("TotalPages(%d, %d) = %d; want %d", tc.total, tc.size, got, tc.want)
		}
	}
}

func TestOffset(t *testing.T) {
	tests := []struct {
		page, size, want int
	}{
		{1, 10, 0},
		{3, 10, 20},
		{0, 10, 0}, // clamps to page 1
	}
	for _, tc := range tests {
		got := Offset(tc.page, tc.size)
		if got != tc.want {
			t.Errorf("Offset(%d, %d) = %d; want %d", tc.page, tc.size, got, tc.want)
		}
	}
}
