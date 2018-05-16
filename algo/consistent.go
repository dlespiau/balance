package algo

import (
	"hash/crc32"
	"sort"
	"strconv"
	"sync"
)

const (
	defaultReplicationCount = 256
)

// ConsistentConfig holds the configuration for the Consistent hash algorithm.
type ConsistentConfig struct {
	// Hash is the hashing function used for hash Endpoints and keys onto the hash
	// ring. You may want to use an interesting hash function like xxHash.
	// Defaults to CRC32.
	Hash Hash
	// ReplicationCount controls the number of virtual nodes to add to the hash
	// ring for each Endpoint.
	// Defaults to 128.
	ReplicationCount int
}

// Consistent implements a consistent hashing algorithm.
type Consistent struct {
	sync.RWMutex
	hash      Hash
	replicas  int
	keys      []int            // Sorted
	endpoints map[int]Endpoint // hash(Endpoint.Key()) -> Endpoint
}

var _ AffinityLoadBalancer = &Consistent{}
var _ EndpointSet = &Consistent{}

// NewConsistent creates a new Consistent object.
func NewConsistent(config ConsistentConfig) *Consistent {
	c := &Consistent{
		replicas:  config.ReplicationCount,
		hash:      config.Hash,
		endpoints: make(map[int]Endpoint),
	}
	if c.replicas == 0 {
		c.replicas = defaultReplicationCount
	}
	if c.hash == nil {
		c.hash = crc32.ChecksumIEEE
	}
	return c
}

// isEmpty returns true if there are no items available.
func (c *Consistent) isEmpty() bool {
	return len(c.keys) == 0
}

// Compute the hash of the ith replica.
func (c *Consistent) replicaHash(key string, i int) int {
	return int(c.hash([]byte(strconv.Itoa(i) + key)))
}

// AddEndpoints implements EndpointSet
func (c *Consistent) AddEndpoints(endpoints ...Endpoint) {
	c.Lock()

	for _, endpoint := range endpoints {
		key := endpoint.Key()
		for i := 0; i < c.replicas; i++ {
			hash := c.replicaHash(key, i)
			c.keys = append(c.keys, hash)
			c.endpoints[hash] = endpoint
		}
	}

	sort.Ints(c.keys)

	c.Unlock()
}

func deleteFromSlice(s []int, hash int) []int {
	var i int

	for i = range s {
		if s[i] == hash {
			break
		}
	}

	if i != len(s) {
		s[i] = s[len(s)-1]
		s = s[:len(s)-1]
	}

	return s
}

// RemoveEndpoints implements EndpointSet
func (c *Consistent) RemoveEndpoints(endpoints ...Endpoint) {
	c.Lock()

	for _, endpoint := range endpoints {
		key := endpoint.Key()
		// XXX: can we do better then O(replicas^2 * endpoints) in the deletion code?
		for i := 0; i < c.replicas; i++ {
			hash := c.replicaHash(key, i)
			c.keys = deleteFromSlice(c.keys, hash)
			delete(c.endpoints, hash)
		}
	}

	sort.Ints(c.keys)

	c.Unlock()
}

// Get implements AffinityLoadBalancer.
func (c *Consistent) Get(key string) Endpoint {
	c.RLock()
	defer c.RUnlock()

	if c.isEmpty() {
		return nil
	}

	hash := int(c.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(c.keys), func(i int) bool { return c.keys[i] >= hash })

	// Means we have cycled back to the first replica.
	if idx == len(c.keys) {
		idx = 0
	}

	return c.endpoints[c.keys[idx]]
}

// Put implements AffinityLoadBalancer.
func (c *Consistent) Put(endpoint Endpoint) {

}
