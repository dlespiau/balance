package e2e

import (
	"time"

	"github.com/dlespiau/balance/e2e/harness"

	appsv1 "k8s.io/api/apps/v1beta2"
)

type fixtures struct {
	test    *harness.Test
	proxy   *appsv1.Deployment
	service *appsv1.Deployment
}

func makeFixtures(test *harness.Test) *fixtures {
	// balance-service Deployment.
	service := test.CreateDeploymentFromFile(test.Namespace, "service-deploy.yaml")

	// balance-proxy Deployment, making sure we tell it to watch the right service.
	proxy := test.LoadDeployment("proxy-deploy.yaml")
	proxy.Spec.Template.Spec.Containers[0].Args = proxyArgs(nil).
		withNamespace(test.Namespace).
		withServiceName(service.Name)
	test.CreateDeployment(test.Namespace, proxy)

	// Create the front-facing Service for balance-service
	serviceService := test.CreateServiceFromFile(test.Namespace, "service-svc.yaml")

	// Wait for things to be ready.
	test.WaitForDeploymentReady(service, 1*time.Minute)
	test.WaitForDeploymentReady(proxy, 1*time.Minute)
	test.WaitForServiceReady(serviceService)

	return &fixtures{
		test:    test,
		proxy:   proxy,
		service: service,
	}
}

func (f *fixtures) sendRequest(key string) ([]byte, error) {
	f.test.Debugf("sending request to the proxy with key %s", key)

	proxyPod := f.test.ListPodsFromDeployment(f.proxy).Items[0]
	return f.test.PodProxyGet(&proxyPod, "", "/hostname").
		SetHeader("X-Affinity", key).
		DoRaw()
}

func (f *fixtures) getServiceStats() []ServiceStats {
	var stats []ServiceStats

	for _, pod := range f.test.ListPodsFromDeployment(f.service).Items {
		podStats := ServiceStats{}
		f.test.PodProxyGetJSON(&pod, "", "/info", &podStats)
		stats = append(stats, podStats)
	}

	return stats
}
