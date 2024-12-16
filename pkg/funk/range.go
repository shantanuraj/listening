package funk

// Range returns a slice of the given slice from start to end.
func Range[A any](xs []A, start int, end int) []A {
	if len(xs) == 0 {
		return []A{}
	}
	if start < 0 {
		start = 0
	}
	if start > len(xs) {
		return []A{}
	}
	if end < 0 {
		end = len(xs) + end
	}
	if end > len(xs) {
		return xs[start:]
	}
	return xs[start:end]
}
