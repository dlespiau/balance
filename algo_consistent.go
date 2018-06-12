package balance

import (
	"hash/crc32"
	"math"
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

	// LoadFactor controls the maximum load of any endpoint. The load is defined as
	// the number of requests currently being handled by an endpoint. When set to a
	// value > 1.0, Consistent implements the bounded loads variant of consistent
	// hashing and ensures no endpoint has a load > LoadFactor * averageLoad.
	//
	// See https://arxiv.org/abs/1608.01350 for details about consistent hashing
	// with bounded loads.
	LoadFactor float64
}

// Store per-endpoint information.
type endpointInfo struct {
	endpoint Endpoint
	load     int
}

// Consistent implements a consistent hashing algorithm.
type Consistent struct {
	sync.Mutex
	hash         Hash
	replicas     int
	numEndpoints int
	loadFactor   float64
	totalLoad    int                   // Total number of requests in flight.
	keys         []int                 // Sorted
	endpoints    map[int]*endpointInfo // hash(Endpoint.Key()) -> endpointInfo
}

var _ LoadBalancer = &Consistent{}
var _ EndpointSet = &Consistent{}

// NewConsistent creates a new Consistent object.
func NewConsistent(config ConsistentConfig) *Consistent {
	c := &Consistent{
		replicas:   config.ReplicationCount,
		loadFactor: config.LoadFactor,
		hash:       config.Hash,
		endpoints:  make(map[int]*endpointInfo),
	}
	// LoadFactor must be > 1.0.
	if c.loadFactor != 0 && c.loadFactor <= 1.0 {
		return nil
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

// info returns the endpointInfo structure for the given endpoint key.
func (c *Consistent) info(key string) *endpointInfo {
	if info, ok := c.endpoints[c.replicaHash(key, 0)]; ok {
		return info
	}
	return nil
}

// AddEndpoints implements EndpointSet
func (c *Consistent) AddEndpoints(endpoints ...Endpoint) {
	c.Lock()

	for _, endpoint := range endpoints {
		key := endpoint.Key()
		info := &endpointInfo{
			endpoint: endpoint,
		}

		for i := 0; i < c.replicas; i++ {
			hash := c.replicaHash(key, i)
			c.keys = append(c.keys, hash)
			c.endpoints[hash] = info
		}

		c.numEndpoints++
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
		info := c.info(key)
		if info == nil {
			continue
		}

		// Update load.
		c.totalLoad -= info.load

		// XXX: can we do better then O(replicas^2 * endpoints) in the deletion code?
		for i := 0; i < c.replicas; i++ {
			hash := c.replicaHash(key, i)
			c.keys = deleteFromSlice(c.keys, hash)
			delete(c.endpoints, hash)
		}

		c.numEndpoints--
	}

	sort.Ints(c.keys)

	c.Unlock()
}

func loadOK(totalLoad, numEndpoints, endpointLoad int, factor float64) bool {
	// We want to ensure the invariant:
	//  endpointLoad <= c * averageLoad
	// -> count the incoming request in the total and endpoint load.
	averageLoad := float64(totalLoad+1) / float64(numEndpoints)
	if (float64(endpointLoad) + 1) <= math.Ceil(factor*averageLoad) {
		return true
	}
	return false
}

// Get implements LoadBalancer.
func (c *Consistent) Get(keys ...string) Endpoint {
	if len(keys) != 1 {
		panic("consistent: affinity key not provided")
	}
	key := keys[0]

	c.Lock()
	defer c.Unlock()

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

	// No bounded loads, simple consistent hashing.
	info := c.endpoints[c.keys[idx]]
	if c.loadFactor == 0 {
		return info.endpoint
	}

	// Search for an endpoint with an acceptable load.
	for {
		if loadOK(c.totalLoad, c.numEndpoints, info.load, c.loadFactor) {
			break
		}

		// Next host, cycling if needed.
		idx++
		if idx >= len(c.keys) {
			idx = 0
		}
		info = c.endpoints[c.keys[idx]]
	}

	// Endpoint found, update load.
	info.load++
	c.totalLoad++

	return info.endpoint
}

// Put implements LoadBalancer.
func (c *Consistent) Put(endpoint Endpoint) {
	c.Lock()
	defer c.Unlock()

	if c.loadFactor == 0 {
		return
	}

	// Update load.
	info := c.info(endpoint.Key())
	if info == nil {
		return
	}
	info.load--
	c.totalLoad--
}
