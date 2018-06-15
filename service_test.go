package balance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSplitHostPort(t *testing.T) {
	tests := []struct {
		input string
		host  string
		port  string
	}{
		{"localhost:http", "localhost", "http"},
		{"localhost:80", "localhost", "80"},
		{"localhost", "localhost", ""},
		{"localhost:", "localhost", ""},
	}

	for _, test := range tests {
		host, port := splitHostPort(test.input)
		assert.Equal(t, test.host, host)
		assert.Equal(t, test.port, port)
	}
}

func TestNewServiceFromString(t *testing.T) {
	tests := []struct {
		input    string
		valid    bool
		expected *Service
	}{
		{"foo.ns", true, &Service{Namespace: "ns", Name: "foo"}},
		{"bar.ns.svc.cluster.local", true, &Service{Namespace: "ns", Name: "bar"}},
		{"bar.ns.svc.cluster.local.", true, &Service{Namespace: "ns", Name: "bar"}},
		{"bar.ns.svc.cluster.local.:8080", true, &Service{Namespace: "ns", Name: "bar", Port: "8080"}},
		{"bar.ns.svc.cluster.local:http", true, &Service{Namespace: "ns", Name: "bar", Port: "http"}},

		{"foo", false, nil},
	}

	for _, test := range tests {
		s, err := NewServiceFromString(test.input)
		if !test.valid {
			assert.Error(t, err)
			assert.Nil(t, s)
			continue
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expected, s)
	}
}
