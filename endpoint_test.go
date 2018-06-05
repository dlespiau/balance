package balance

// e dummy implementation of Endpoint for testing.
type e string

func (e e) Key() string { return string(e) }

// el is a convenience function to build a list of endpoints from a list of strings
func el(l ...string) []Endpoint {
	endpoints := make([]Endpoint, 0, len(l))
	for _, s := range l {
		endpoints = append(endpoints, e(s))
	}

	return endpoints
}

func keys(endpoints []Endpoint) []string {
	keys := make([]string, 0, len(endpoints))
	for _, e := range endpoints {
		keys = append(keys, e.Key())
	}
	return keys
}
