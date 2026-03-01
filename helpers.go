package table

// pageSizes returns the page sizes list, falling back to defaults if empty.
func pageSizes(sizes []int) []int {
	if len(sizes) > 0 {
		return sizes
	}
	return DefaultPageSizes()
}
