package table

import "testing"

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
	u := SortURL("/items?page=3", "name", "", "")
	if u == "" {
		t.Fatal("SortURL returned empty")
	}
	// Should not contain page param
	if got := u; got == "/items?page=3&sort_by=name&sort_dir=asc" {
		t.Error("SortURL should reset page param")
	}
}

func TestPageURL(t *testing.T) {
	u := PageURL("/items?sort_by=name", 5)
	want := "/items?page=5&sort_by=name"
	if u != want {
		t.Errorf("PageURL() = %q; want %q", u, want)
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
	}
	for _, tc := range tests {
		got := PaginationInfo(tc.page, tc.pageSize, tc.total)
		if got != tc.want {
			t.Errorf("PaginationInfo(%d,%d,%d) = %q; want %q",
				tc.page, tc.pageSize, tc.total, got, tc.want)
		}
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
	if len(got) != 2 || got[0] != 5 {
		t.Errorf("pageSizes(custom) = %v; want [5,15]", got)
	}
}
