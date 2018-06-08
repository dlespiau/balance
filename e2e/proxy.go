package e2e

type proxyArgs []string

func (args proxyArgs) withServiceName(serviceName string) proxyArgs {
	return append(args, "-k8s.service="+serviceName)
}

func (args proxyArgs) withNamespace(ns string) proxyArgs {
	return append(args, "-k8s.namespace="+ns)
}
