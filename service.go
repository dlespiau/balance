package balance

import (
	"fmt"
	"strings"
)

// Service describes a Kubernetes service. Port can be a named port or the port
// number.
type Service struct {
	Namespace, Name, Port string
}

// validOptionalPort reports whether port is either an empty string
// or matches /^:\d*$/
func validOptionalPort(port string) bool {
	if port == "" {
		return true
	}
	if port[0] != ':' {
		return false
	}
	for _, b := range port[1:] {
		if b < '0' || b > '9' {
			return false
		}
	}
	return true
}

// splitHostPort splits host and port from "host:port" strings.
func splitHostPort(hostport string) (host, port string) {
	colon := strings.LastIndex(hostport, ":")
	if colon < 0 {
		return hostport, ""
	}
	return hostport[:colon], hostport[colon+1:]
}

// NewServiceFromString parses the kubernetes Service hostname and port and
// returns a Service object.
func NewServiceFromString(s string) (*Service, error) {
	host, port := splitHostPort(s)

	// XXX suport only one element in the host, defaulting to the namespace we are
	// running in.
	parts := strings.SplitN(host, ".", 3)
	if len(parts) < 2 {
		return nil, fmt.Errorf("service: expected name and namespace, got %s", s)
	}

	return &Service{
		Namespace: parts[1],
		Name:      parts[0],
		Port:      port,
	}, nil
}

func (s *Service) String() string {
	host := s.Name + "." + s.Namespace
	if s.Port != "" {
		host += ":" + s.Port
	}
	return host
}
