package balance

// kubernetesEndpoint is a Kubernetes Service endpoint.
type kubernetesEndpoint struct {
	Address string
}

var _ Endpoint = &kubernetesEndpoint{}

// Key implements Endpoint.
func (e *kubernetesEndpoint) Key() string {
	return e.Address
}

// String implements fmt.Stringer.
func (e *kubernetesEndpoint) String() string {
	return e.Address
}
