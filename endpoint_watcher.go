package balance

import (
	"fmt"
	"net"
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	watch "k8s.io/apimachinery/pkg/watch"
	client "k8s.io/client-go/kubernetes"
)

// EndpointWatcher load balances requests.
type EndpointWatcher struct {
	Client   *client.Clientset
	Service  Service
	Receiver EndpointSet

	previousEndpoints []Endpoint
}

func (w *EndpointWatcher) makeEndpoints(subsets []corev1.EndpointSubset) []Endpoint {
	var endpoints []Endpoint

	servicePort, err := strconv.Atoi(w.Service.Port)

	for _, subset := range subsets {
		for _, address := range subset.Addresses {
			for _, port := range subset.Ports {
				if (err == nil && servicePort == int(port.Port)) ||
					(err != nil && w.Service.Port == port.Name) {
					endpoints = append(endpoints, &kubernetesEndpoint{
						Address: net.JoinHostPort(address.IP, strconv.Itoa(int(port.Port))),
					})
				}
			}
		}
	}

	return endpoints
}

func (w *EndpointWatcher) setEndpoints(endpoints []Endpoint) {
	for _, chunk := range difference(w.previousEndpoints, endpoints) {
		switch chunk.operation {
		case add:
			log.Debugf("watcher: add %s", chunk.endpoint.Key())
			w.Receiver.AddEndpoints(chunk.endpoint)
		case del:
			log.Debugf("watcher: remove %s", chunk.endpoint.Key())
			w.Receiver.RemoveEndpoints(chunk.endpoint)
		}
	}

	w.previousEndpoints = endpoints
}

func (w *EndpointWatcher) watchEvents(events watch.Interface, close <-chan interface{}) error {
	select {
	case <-close:
		log.Info("stop watching endpoints")
		return nil
	case event := <-events.ResultChan():
		switch event.Type {
		case watch.Added:
			fallthrough
		case watch.Modified:
			endpoints, ok := event.Object.(*corev1.Endpoints)
			if !ok {
				// Should never happen as it'd break the documented API contract.
				return fmt.Errorf("could not cast a %s object to Endpoints", event.Object.GetObjectKind().GroupVersionKind().Kind)
			}
			newEndpoints := w.makeEndpoints(endpoints.Subsets)
			w.setEndpoints(newEndpoints)
		case watch.Deleted:
			w.setEndpoints(nil)
		case watch.Error:
			status, ok := event.Object.(*metav1.Status)
			if !ok {
				// Should never happen as it'd break the documented API contract.
				return fmt.Errorf("could not cast a %s object to Status", event.Object.GetObjectKind().GroupVersionKind().Kind)
			}
			return errors.FromObject(status)
		}
	}

	return nil
}

// Start will start an internal goroutine that watches the kubernetes service
// and notify the Receiver of Endpoints changes.
// close is a channel the caller can close to terminate this goroutine.
func (w *EndpointWatcher) Start(close <-chan interface{}) {
	go func() {
		for {
			events, err := w.Client.CoreV1().Endpoints(w.Service.Namespace).Watch(metav1.ListOptions{
				FieldSelector: fmt.Sprintf("metadata.name=%s", w.Service.Name),
			})
			if err != nil {
				log.Fatal(err)
				return
			}

			err = w.watchEvents(events, close)
			events.Stop()
			if err != nil {
				log.Error(err)
				// XXX we may need to reconnect the Kubernetes client.
				time.Sleep(time.Second)
			}
		}
	}()
}
