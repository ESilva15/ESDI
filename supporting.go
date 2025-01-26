package main

func findEntry[T any](s []T, predicate func(T) bool) int {
	for i, elem := range s {
		if predicate(elem) {
			return i
		}
	}

	return -1
}
