package balance

// e dummy implementation of Endpoint for testing.
type e string

func (e e) Key() string { return string(e) }
