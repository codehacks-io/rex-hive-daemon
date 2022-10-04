package slice_tools

// FindIndex returns the first index satisfying the given function, or -1 if none do.
func FindIndex[E any](slice *[]E, fn func(E) bool) int {
	for i, v := range *slice {
		if fn(v) {
			return i
		}
	}
	return -1
}

// RemoveAtIndex returns a slice without the element at the given index.
func RemoveAtIndex[E any](slice *[]E, index int) []E {
	return append((*slice)[:index], (*slice)[index+1:]...)
}
