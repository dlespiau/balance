package balance

type differenceOperation int

const (
	add differenceOperation = 0
	del differenceOperation = 1
)

type chunk struct {
	operation differenceOperation
	endpoint  Endpoint
}

func isIn(needle Endpoint, haystack []Endpoint) bool {
	for i := range haystack {
		if haystack[i].Key() == needle.Key() {
			return true
		}
	}
	return false
}

// diff is a - b, the classical set difference.
func diff(a, b []Endpoint) []Endpoint {
	var result []Endpoint

	// We assume the slices are small enough that a O(1) check with maps isn't an
	// optimization, nor is sorting the slices for early return.
	for i := range a {
		if isIn(a[i], b) {
			continue
		}
		result = append(result, a[i])
	}

	return result
}

func difference(old, new []Endpoint) []chunk {
	var chunks []chunk

	for _, e := range diff(old, new) {
		chunks = append(chunks, chunk{del, e})
	}
	for _, e := range diff(new, old) {
		chunks = append(chunks, chunk{add, e})
	}

	return chunks
}
