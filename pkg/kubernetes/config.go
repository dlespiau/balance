package kubernetes

import (
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	client "k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

// ClientConfig holds configuration options for API server clients.
type ClientConfig struct {
	CertificateAuthority string
	ClientCertificate    string
	ClientKey            string
	Cluster              string
	Context              string
	Insecure             bool
	Kubeconfig           string
	Password             string
	Server               string
	Token                string
	User                 string
	Username             string
}

func homeDirectory() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}

// kubeconfigPath returns the default kubeconfig location.
func kubeconfigPath() string {
	home := homeDirectory()
	if home == "" {
		return ""
	}
	return filepath.Join(home, ".kube", "config")
}

// NewClientConfig returns a configuration object that can be used to configure a client in
// order to contact an API server with.
func NewClientConfig(config *ClientConfig) (*rest.Config, error) {
	var restConfig *rest.Config
	var err error

	if config.Server == "" && config.Kubeconfig == "" {
		// If no API server address or kubeconfig was provided, assume we are
		// running inside a pod and Try to connect to the API server through
		// its Service environment variables, using the default Service Account
		// Token.
		restConfig, err = rest.InClusterConfig()
	}

	if restConfig == nil {
		// We're not in a pod? try to use kubeconfig.

		// Try the default kubeconfig location if nothing else is provided.
		if config.Kubeconfig == "" {
			config.Kubeconfig = kubeconfigPath()
		}

		restConfig, err = clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
			&clientcmd.ClientConfigLoadingRules{ExplicitPath: config.Kubeconfig},
			&clientcmd.ConfigOverrides{
				AuthInfo: clientcmdapi.AuthInfo{
					ClientCertificate: config.ClientCertificate,
					ClientKey:         config.ClientKey,
					Token:             config.Token,
					Username:          config.Username,
					Password:          config.Password,
				},
				ClusterInfo: clientcmdapi.Cluster{
					Server:                config.Server,
					InsecureSkipTLSVerify: config.Insecure,
					CertificateAuthority:  config.CertificateAuthority,
				},
				Context: clientcmdapi.Context{
					Cluster:  config.Cluster,
					AuthInfo: config.User,
				},
				CurrentContext: config.Context,
			},
		).ClientConfig()
	}

	if err != nil {
		return nil, err
	}

	log.Infof("kubernetes: targeting api server %s", restConfig.Host)

	return restConfig, nil
}

// NewClientWithConfig will create a new client able to talk to a Kubernetes API
// server.
func NewClientWithConfig(config *ClientConfig) (*client.Clientset, error) {
	kubeConfig, err := NewClientConfig(config)
	if err != nil {
		return nil, err
	}

	return client.NewForConfig(kubeConfig)
}
