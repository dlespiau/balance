package balance

type baseMiddleware struct {
	next Algorithm
}

var _ Algorithm = &baseMiddleware{}

func (m *baseMiddleware) AddEndpoints(endpoints ...Endpoint) {
	m.next.AddEndpoints(endpoints...)
}

func (m *baseMiddleware) RemoveEndpoints(endpoints ...Endpoint) {
	m.next.RemoveEndpoints(endpoints...)
}

func (m *baseMiddleware) Get(key ...string) Endpoint {
	return m.next.Get(key...)
}

func (m *baseMiddleware) Put(endpoint Endpoint) {
	m.next.Put(endpoint)
}
