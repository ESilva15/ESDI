package main

func findEntry[T any](s []T, predicate func(T) bool) int {
	for i, elem := range s {
		if predicate(elem) {
			return i
		}
	}

	return -1
}

func copyBytes(dest []byte, destSize int, src string) {
	copy(dest[:], []byte(src))
	dest[min(destSize-1, len(src))] = '\x00'
}
