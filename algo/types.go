package algo

// Hash is a 32-bit hash function.
type Hash func(data []byte) uint32

// Endpoint is service endpoint.
type Endpoint interface {
	Key() string
}

// EndpointSet holds a set of Endpoints.
type EndpointSet interface {
	AddEndpoints(...Endpoint)
	RemoveEndpoints(...Endpoint)
}

// LoadBalancer is an interface abstracting load balancing.
type LoadBalancer interface {
	// Get returns the Service Endpoint to use for the next request. Get returns
	// nil when no Endpoint hash been added to the load balancer.
	Get() Endpoint
	// Put releases the Endpoint when it has finished processing the request.
	Put(endpoint Endpoint)
}

// AffinityLoadBalancer is an interface abstracting load balancing algorithms
// with an affinity scheme.
type AffinityLoadBalancer interface {
	// Get is called when wanting to send a request to a Service. It returns the
	// Endpoint that request should be directed to, based on the affinity given
	// affinity key. Get returns nil when no Endpoint hash been added to the load
	// balancer.
	Get(key string) Endpoint
	// Put releases the Endpoint when it has finished processing the request.
	Put(endpoint Endpoint)
}
