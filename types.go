package balance

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

// LoadBalancer is an interface abstracting load balancing algorithms.
type LoadBalancer interface {
	EndpointSet
	// Get returns the Service Endpoint to use for the next request.
	//
	// Get is called when wanting to send a request to a Service. It returns the
	// Endpoint that request should be directed to. When used with an affinity load
	// balancing scheme, the affinity key needs to be given as argument to this
	// function.
	//
	// When used with an affinity load balancing scheme, Get will panic if no key
	// or more than one key is given. The variadic form is only used for
	// aesthetics, ie. being able to use Get() which non-affinity load balancers.
	//
	// Get may return nil when the load balancer is not aware of any Endpoint. It's
	// possible to tweak this behavior by wrapping a LoadBalancer into
	Get(key ...string) Endpoint
	// Put releases the Endpoint when it has finished processing the request.
	Put(endpoint Endpoint)
}
