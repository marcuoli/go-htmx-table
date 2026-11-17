package table

import "net/url"

// pageSizes returns the page sizes list, falling back to defaults if empty.
func pageSizes(sizes []int) []int {
	if len(sizes) > 0 {
		return sizes
	}
	return DefaultPageSizes()
}

// HTMX appends the selected size, so existing pagination keys must be absent.
func pageSizeBaseURL(baseURL string) string {
	u, err := url.Parse(baseURL)
	if err != nil {
		return baseURL
	}
	q := u.Query()
	q.Del("page")
	q.Del("page_size")
	u.RawQuery = q.Encode()
	return u.String()
}
