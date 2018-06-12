package e2e

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConsistentAllRequestsServed(t *testing.T) {
	t.Parallel()

	test := kube.NewTest(t).Setup()
	defer test.Close()

	f := makeFixtures(test)

	// Send 50 requests to the proxy, different key each time.
	const numSentRequests = 50
	test.Infof("sending %d requests to the proxy", numSentRequests)
	for i := 0; i < numSentRequests; i++ {
		_, err := f.sendRequest(fmt.Sprintf("%d", i))
		assert.NoError(t, err)
	}

	// Ensure all requests have been received.
	numReceivedRequests := 0
	for _, stats := range f.getServiceStats() {
		numReceivedRequests += stats.RequestCount
	}
	assert.Equal(t, numSentRequests, numReceivedRequests)
}
