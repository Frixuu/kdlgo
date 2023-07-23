package kdl

type pair[K comparable, V any] struct {
	key   K
	value V
}

// toPairs converts a `map[K]V` to a slice of `pair[K, V]` structs.
func toPairs[K comparable, V any](m map[K]V) []pair[K, V] {
	var pairs []pair[K, V] = make([]pair[K, V], 0, len(m))
	for k, v := range m {
		pairs = append(pairs, pair[K, V]{key: k, value: v})
	}
	return pairs
}
