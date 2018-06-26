package balance

type serviceFallback struct {
	baseMiddleware
	service Service
}

var _ Algorithm = &serviceFallback{}

// WithServiceFallback wraps a load balancer, falling back to the service DNS
// name when there's no available endpoint to serve the request.
func WithServiceFallback(next Algorithm, service *Service) Algorithm {
	return &serviceFallback{
		baseMiddleware: baseMiddleware{
			next: next,
		},
		service: *service,
	}
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
