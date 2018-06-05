package balance

// Service describes a Kubernetes service. Port can be a named port or the port
// number.
type Service struct {
	Namespace, Name, Port string
}
