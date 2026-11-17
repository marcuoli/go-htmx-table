package table

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"math"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/a-h/templ"
	"github.com/marcuoli/go-htmx-table/queryutil"
	"golang.org/x/net/html"
)

func TestTableID_Default(t *testing.T) {
	p := DataTableProps{}
	if got := p.TableID(); got != "data-table" {
		t.Errorf("TableID() = %q; want %q", got, "data-table")
	}
}

func TestTableID_Custom(t *testing.T) {
	p := DataTableProps{ID: "my-table"}
	if got := p.TableID(); got != "my-table" {
		t.Errorf("TableID() = %q; want %q", got, "my-table")
	}
}

func TestDefaultPageSizes(t *testing.T) {
	sizes := DefaultPageSizes()
	want := []int{10, 20, 50, 100}
	if len(sizes) != len(want) {
		t.Fatalf("DefaultPageSizes() length = %d; want %d", len(sizes), len(want))
	}
	for i, v := range want {
		if sizes[i] != v {
			t.Errorf("DefaultPageSizes()[%d] = %d; want %d", i, sizes[i], v)
		}
	}
}

func TestSortURL_Toggle(t *testing.T) {
	// Currently ascending on "name" → should flip to descending
	u := SortURL("/items", "name", "name", "asc")
	if u != "/items?sort_by=name&sort_dir=desc" {
		t.Errorf("SortURL() = %q; want descending", u)
	}
}

func TestSortURL_NewField(t *testing.T) {
	// Different field → should default to ascending
	u := SortURL("/items", "email", "name", "asc")
	if u != "/items?sort_by=email&sort_dir=asc" {
		t.Errorf("SortURL() = %q; want ascending on new field", u)
	}
}

func TestSortURL_ResetsPage(t *testing.T) {
	got := SortURL("/items?page=3&status=active", "name", "", "")
	want := "/items?sort_by=name&sort_dir=asc&status=active"
	if got != want {
		t.Errorf("SortURL() = %q; want %q", got, want)
	}
}

func TestPageURL(t *testing.T) {
	u := PageURL("/items?sort_by=name", 5)
	want := "/items?page=5&sort_by=name"
	if u != want {
		t.Errorf("PageURL() = %q; want %q", u, want)
	}
}

func TestListProps_StateURL_OmitsSessionDerivedPageSize(t *testing.T) {
	p := ListProps{BaseURL: "/items", PageSize: 15, Search: "abc"}
	if got := p.StateURL(); got != "/items?search=abc" {
		t.Errorf("StateURL() = %q; want no page_size", got)
	}
}

func TestListProps_StateURL_PreservesExplicitPageSize(t *testing.T) {
	p := ListProps{BaseURL: "/items", PageSize: 50, PageSizeExplicit: true, Search: "abc"}
	if got := p.StateURL(); got != "/items?page_size=50&search=abc" {
		t.Errorf("StateURL() = %q; want explicit page_size", got)
	}
}

func TestListProps_PageSizeBaseURL(t *testing.T) {
	p := ListProps{
		BaseURL: "/items?page=3&page=4&page_size=10&page_size=20&status=active&status=pending#results",
		Search:  "a & b", SearchName: "q", SortBy: "name", SortDir: "desc",
		PageSize: 10, PageSizeExplicit: true,
	}
	want := "/items?q=a+%26+b&sort_by=name&sort_dir=desc&status=active&status=pending#results"
	if got := p.PageSizeBaseURL(); got != want {
		t.Errorf("PageSizeBaseURL() = %q; want %q", got, want)
	}
}

func TestPageSizeSelectors(t *testing.T) {
	baseURL := "/items?page=3&page_size=10&page_size=20&search=abc&sort_by=name&sort_dir=desc&status=active#results"
	for _, tc := range []struct {
		name      string
		component templ.Component
	}{
		{"pagination", ServerPagination(PaginationProps{
			BaseURL: baseURL, Page: 3, PageSize: 10, TotalPages: 5, TotalItems: 42,
			HxTarget: "#results", HxPushURL: true,
		})},
		{"list", List(ListProps{
			BaseURL: baseURL, Page: 3, PageSize: 10, PageSizeExplicit: true,
			TotalPages: 5, TotalItems: 42, Search: "abc", SortBy: "name", SortDir: "desc",
			HxTarget: "#results", HxPushURL: true,
		})},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var rendered bytes.Buffer
			if err := tc.component.Render(context.Background(), &rendered); err != nil {
				t.Fatal(err)
			}
			attrs := pageSizeSelectAttributes(t, &rendered)
			for name, want := range map[string]string{"hx-target": "#results", "hx-push-url": "true", "hx-include": "this"} {
				if attrs[name] != want {
					t.Errorf("%s = %q; want %q", name, attrs[name], want)
				}
			}
			u, err := url.Parse(attrs["hx-get"])
			if err != nil {
				t.Fatal(err)
			}
			wantURL := "/items?search=abc&sort_by=name&sort_dir=desc&status=active#results"
			if u.String() != wantURL {
				t.Errorf("selector URL = %q; want %q", u, wantURL)
			}
			// HTMX appends the control value to the URL's existing query.
			q := u.Query()
			q.Add("page_size", "50")
			u.RawQuery = q.Encode()
			r := httptest.NewRequest("GET", u.String(), nil)
			if got := queryutil.ParsePageSize(r, 10, 100); got != 50 {
				t.Errorf("selected page size = %d; want 50", got)
			}
			if got := queryutil.ParsePage(r); got != 1 {
				t.Errorf("page after size change = %d; want 1", got)
			}
		})
	}
}

func pageSizeSelectAttributes(t *testing.T, rendered io.Reader) map[string]string {
	t.Helper()
	tokens := html.NewTokenizer(rendered)
	for {
		switch tokens.Next() {
		case html.ErrorToken:
			t.Fatalf("page_size select not found: %v", tokens.Err())
		case html.StartTagToken:
			token := tokens.Token()
			if token.Data != "select" {
				continue
			}
			attrs := make(map[string]string, len(token.Attr))
			for _, attr := range token.Attr {
				attrs[attr.Key] = attr.Val
			}
			if attrs["name"] == "page_size" {
				return attrs
			}
		}
	}
}

func TestPageSizeURL(t *testing.T) {
	u := PageSizeURL("/items?page=3", 50)
	// Should set page_size and remove page
	want := "/items?page_size=50"
	if u != want {
		t.Errorf("PageSizeURL() = %q; want %q", u, want)
	}
}

func TestSortIcon_Active_Asc(t *testing.T) {
	if got := SortIcon("name", "name", "asc"); got != "▲" {
		t.Errorf("SortIcon() = %q; want ▲", got)
	}
}

func TestSortIcon_Active_Desc(t *testing.T) {
	if got := SortIcon("name", "name", "desc"); got != "▼" {
		t.Errorf("SortIcon() = %q; want ▼", got)
	}
}

func TestSortIcon_Inactive(t *testing.T) {
	if got := SortIcon("email", "name", "asc"); got != "↕" {
		t.Errorf("SortIcon() = %q; want ↕", got)
	}
}

func TestPaginationInfo(t *testing.T) {
	tests := []struct {
		page, pageSize, total int
		want                  string
	}{
		{1, 10, 42, "Showing 1\u201310 of 42 items"},
		{5, 10, 42, "Showing 41\u201342 of 42 items"},
		{1, 10, 0, "No items"},
		{1, 20, 5, "Showing 1\u20135 of 5 items"},
		{0, 10, 42, "Showing 1\u201310 of 42 items"},
		{math.MaxInt, 10, 42, "Showing 41\u201342 of 42 items"},
		{1, 0, 42, "No items"},
		{1, 10, -1, "No items"},
	}
	for _, tc := range tests {
		got := PaginationInfo(tc.page, tc.pageSize, tc.total)
		if got != tc.want {
			t.Errorf("PaginationInfo(%d,%d,%d) = %q; want %q",
				tc.page, tc.pageSize, tc.total, got, tc.want)
		}
	}
}

func TestPaginationInfo_IntegerLimit(t *testing.T) {
	want := fmt.Sprintf("Showing %d\u2013%d of %d items", math.MaxInt/2+2, math.MaxInt, math.MaxInt)
	if got := PaginationInfo(2, math.MaxInt/2+1, math.MaxInt); got != want {
		t.Errorf("PaginationInfo() = %q; want %q", got, want)
	}
}

func TestBadgeVariantConstants(t *testing.T) {
	// Ensure constants have expected string values
	if BadgeSuccess != "success" {
		t.Errorf("BadgeSuccess = %q", BadgeSuccess)
	}
	if BadgeDanger != "danger" {
		t.Errorf("BadgeDanger = %q", BadgeDanger)
	}
}

func TestVariantFromString(t *testing.T) {
	tests := []struct {
		input string
		want  BadgeVariant
	}{
		{"active", BadgeSuccess},
		{"Active", BadgeSuccess},
		{"completed", BadgeSuccess},
		{"pending", BadgeWarning},
		{"in progress", BadgeWarning},
		{"error", BadgeDanger},
		{"failed", BadgeDanger},
		{"info", BadgeInfo},
		{"draft", BadgeInfo},
		{"unknown", BadgeNeutral},
		{"", BadgeNeutral},
	}
	for _, tc := range tests {
		got := variantFromString(tc.input)
		if got != tc.want {
			t.Errorf("variantFromString(%q) = %q; want %q", tc.input, got, tc.want)
		}
	}
}

func TestPageSizes_Fallback(t *testing.T) {
	got := pageSizes(nil)
	if len(got) != 4 {
		t.Errorf("pageSizes(nil) length = %d; want 4", len(got))
	}
}

func TestPageSizes_Custom(t *testing.T) {
	custom := []int{5, 15}
	got := pageSizes(custom)
	if len(got) != 2 || got[0] != 5 || got[1] != 15 {
		t.Errorf("pageSizes(custom) = %v; want [5,15]", got)
	}
}

func TestListProps_ShouldShowPagination(t *testing.T) {
	tests := []struct {
		name  string
		props ListProps
		want  bool
	}{
		{name: "multiple pages", props: ListProps{TotalPages: 2, TotalItems: 25}, want: true},
		{name: "single page default hidden", props: ListProps{TotalPages: 1, TotalItems: 10}, want: false},
		{name: "single page opt in", props: ListProps{TotalPages: 1, TotalItems: 10, AlwaysShowPagination: true}, want: true},
		{name: "empty opt in hidden", props: ListProps{TotalPages: 0, TotalItems: 0, AlwaysShowPagination: true}, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := tc.props.ShouldShowPagination(); got != tc.want {
				t.Errorf("ShouldShowPagination() = %v; want %v", got, tc.want)
			}
		})
	}
}
