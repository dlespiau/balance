package balance

import (
	"k8s.io/client-go/kubernetes"
)

// LoadBalancerOptions are configuration parameters.
type LoadBalancerOptions struct {
	// Fallback defines the strategy to adopt when the load balancer had no
	// endpoint. If not specified, defaults to FallbackService.
	Fallback Fallback
}

// LoadBalancer is a Kubernetes Service load balancer.
type LoadBalancer struct {
	kubeClient kubernetes.Interface
	service    string
	algo       Algorithm // innermost Algorithm as configured by the user
	balancer   Algorithm // outermost Algorithm
	opts       LoadBalancerOptions
}

// Fallback is a fallback strategy.
type Fallback string

const (
	// FallbackNone disables fallback strategies. Get will return nil when the
	// Service has no endpoint.
	FallbackNone Fallback = "none"
	// FallbackService is a fallback strategy that will make the load balancer use
	// the Service DNS name.
	FallbackService Fallback = "service"
)

// NewLoadBalancer creates a new load balancer for service.
func NewLoadBalancer(service string, algo Algorithm, options LoadBalancerOptions) *LoadBalancer {
	return &LoadBalancer{
		service: service,
		algo:    algo,
		opts:    options,
	}
}

// Start initializes the load balancer. Start must be called before any other
// function.
//
// The load balancer object may create goroutines or other precious resources.
// Stopping a load balancer is done by closing the stop channel.
func (lb *LoadBalancer) Start(stop <-chan interface{}) error {
	service, err := NewServiceFromString(lb.service)
	if err != nil {
		return err
	}

	client, err := makeInClusterClient()
	if err != nil {
		return err
	}

	if lb.opts.Fallback == "" {
		lb.opts.Fallback = FallbackService
	}

	watcher := EndpointWatcher{
		Client:   client,
		Service:  *service,
		Receiver: lb.algo,
	}

	watcher.Start(make(<-chan interface{}))

	lb.balancer = lb.algo

	switch lb.opts.Fallback {
	case FallbackService:
		lb.balancer = WithServiceFallback(lb.balancer, service)
	}

	return nil
}

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
// Get may return nil when the load balancer is not aware of any
// Endpoint.
func (lb *LoadBalancer) Get(key ...string) Endpoint {
	return lb.balancer.Get(key...)
}

// Put releases the Endpoint when it has finished processing the request.
func (lb *LoadBalancer) Put(endpoint Endpoint) {
	lb.balancer.Put(endpoint)
}
