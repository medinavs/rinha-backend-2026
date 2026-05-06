//go:build !linux

package vectorindex

func loadBinaryIndex(path string) (*QuantizedIndex, error) {
	return loadBinaryIndexHeap(path)
}
