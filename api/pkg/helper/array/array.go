package array

// Reverse an array of T mutating the array
func Reverse[T any](array []T) {
	for i, j := 0, len(array)-1; i < j; i, j = i+1, j-1 {
		array[i], array[j] = array[j], array[i]
	}
}

// Chunk splits a source slice into a slice of smaller slices (chunks),
// where each smaller slice has at most chunkSize elements.
// T represents the element type, allowing it to work with any slice (e.g., []int, []string, []*MyStruct).
func Chunk[T any](s []T, chunkSize int) [][]T {
	if chunkSize <= 0 || len(s) == 0 {
		return [][]T{}
	}

	// Calculate the number of chunks needed (ceiling division)
	numChunks := (len(s) + chunkSize - 1) / chunkSize
	result := make([][]T, 0, numChunks)

	// Loop using an index (i) that jumps by the chunkSize in each iteration.
	for i := 0; i < len(s); i += chunkSize {
		end := min(i+chunkSize, len(s))

		result = append(result, s[i:end])
	}

	return result
}
