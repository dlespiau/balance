package e2e

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConsistentAllRequestsServed(t *testing.T) {
	t.Parallel()

	test := kube.NewTest(t).Setup()
	defer test.Close()

	// Deploy service pods and proxy in the cluster.
	service := test.CreateDeploymentFromFile(test.Namespace, "service-deploy.yaml")

	proxy := test.LoadDeployment("proxy-deploy.yaml")
	proxy.Spec.Template.Spec.Containers[0].Args = proxyArgs(nil).
		withNamespace(test.Namespace).
		withServiceName(service.Name)
	test.CreateDeployment(test.Namespace, proxy)

	serviceService := test.CreateServiceFromFile(test.Namespace, "service-svc.yaml")

	test.WaitForDeploymentReady(service, 1*time.Minute)
	test.WaitForDeploymentReady(proxy, 1*time.Minute)
	test.WaitForServiceReady(serviceService)

	// Send 100 requests to the proxy, different key each time.
	const numSentRequests = 100
	test.Infof("sending %d requests to the proxy", numSentRequests)
	proxyPod := test.ListPodsFromDeployment(proxy).Items[0]
	for i := 0; i < numSentRequests; i++ {
		_, err := test.PodProxyGet(&proxyPod, "", "/hostname").SetHeader("X-Affinity", "Bob").DoRaw()
		assert.NoError(t, err)
	}

	numReceivedRequests := 0
	for _, pod := range test.ListPodsFromDeployment(service).Items {
		stats := ServiceStats{}
		test.PodProxyGetJSON(&pod, "", "/info", &stats)
		numReceivedRequests += stats.RequestCount
	}

	assert.Equal(t, numSentRequests, numReceivedRequests)
}
