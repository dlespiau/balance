package balance

type serviceFallback struct {
	next    LoadBalancer
	service Service
}

var _ LoadBalancer = &serviceFallback{}

// WithServiceFallback wraps a load balancer, falling back to the service DNS
// name when there's no available endpoint to serve the request.
func WithServiceFallback(next LoadBalancer, service *Service) LoadBalancer {
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
