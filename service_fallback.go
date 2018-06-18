package balance

type serviceFallback struct {
	next    Algorithm
	service Service
}

var _ Algorithm = &serviceFallback{}

// WithServiceFallback wraps a load balancer, falling back to the service DNS
// name when there's no available endpoint to serve the request.
func WithServiceFallback(next Algorithm, service *Service) Algorithm {
	return &serviceFallback{
		next:    next,
		service: *service,
	}
}

func (sf *serviceFallback) AddEndpoints(endpoints ...Endpoint) {
	sf.next.AddEndpoints(endpoints...)
}

func (sf *serviceFallback) RemoveEndpoints(endpoints ...Endpoint) {
	sf.next.RemoveEndpoints(endpoints...)
}

func (sf *serviceFallback) Get(key ...string) Endpoint {
	endpoint := sf.next.Get(key...)
	if endpoint != nil {
		return endpoint
	}
	return &kubernetesEndpoint{
		Address: sf.service.Name + "." + sf.service.Namespace + ":" + sf.service.Port,
	}
}

func (sf *serviceFallback) Put(endpoint Endpoint) {
	sf.next.Put(endpoint)
}
