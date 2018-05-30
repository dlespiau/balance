// Copyright 2013 Google Inc.
// Copyright 2018 Weaveworks
//
// Originally from:
//  https://github.com/golang/groupcache/blob/master/consistenthash/consistenthash_test.go

package balance

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type e string

func (e e) Key() string { return string(e) }

func testHash(key []byte) uint32 {
	i, err := strconv.Atoi(string(key))
	if err != nil {
		panic(err)
	}
	return uint32(i)
}

func TestHashing(t *testing.T) {

	// Override the hash function to return easier to reason about values. Assumes
	// the keys can be converted to an integer.
	hash := NewConsistent(ConsistentConfig{
		ReplicationCount: 3,
		Hash:             testHash,
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.AddEndpoints(e("6"), e("4"), e("2"))
	assert.Equal(t, []int{2, 4, 6, 12, 14, 16, 22, 24, 26}, hash.keys)
	assert.Equal(t, hash.numEndpoints, 3)

	testCases := map[string]Endpoint{
		"2":  e("2"),
		"11": e("2"),
		"23": e("4"),
		"27": e("2"),
	}

	for k, v := range testCases {
		assert.Equal(t, v, hash.Get(k))
	}

	// Adds 8, 18, 28
	hash.AddEndpoints(e("8"))
	assert.Equal(t, []int{2, 4, 6, 8, 12, 14, 16, 18, 22, 24, 26, 28}, hash.keys)
	assert.Equal(t, hash.numEndpoints, 4)

	// 27 should now map to 8.
	testCases["27"] = e("8")

	for k, v := range testCases {
		assert.Equal(t, v, hash.Get(k))
	}

	// Removes 8, 18, 28
	hash.RemoveEndpoints(e("8"))
	assert.Equal(t, []int{2, 4, 6, 12, 14, 16, 22, 24, 26}, hash.keys)
	assert.Equal(t, hash.numEndpoints, 3)

	// 27 should now map to 2 again.
	testCases["27"] = e("2")

	for k, v := range testCases {
		assert.Equal(t, v, hash.Get(k))
	}
}

func TestConsistency(t *testing.T) {
	hash1 := NewConsistent(ConsistentConfig{ReplicationCount: 1})
	hash2 := NewConsistent(ConsistentConfig{ReplicationCount: 1})

	hash1.AddEndpoints(e("Bill"), e("Bob"), e("Bonny"))
	hash2.AddEndpoints(e("Bob"), e("Bonny"), e("Bill"))

	if hash1.Get("Ben") != hash2.Get("Ben") {
		t.Errorf("Fetching 'Ben' from both hashes should be the same")
	}

	hash2.AddEndpoints(e("Becky"), e("Ben"), e("Bobby"))

	if hash1.Get("Ben") != hash2.Get("Ben") ||
		hash1.Get("Bob") != hash2.Get("Bob") ||
		hash1.Get("Bonny") != hash2.Get("Bonny") {
		t.Errorf("Direct matches should always return the same entry")
	}

}

func TestEndpointDisappearing(t *testing.T) {
	hash := NewConsistent(ConsistentConfig{
		ReplicationCount: 3,
		Hash:             testHash,
		LoadFactor:       1.25,
	})

	hash.AddEndpoints(e("6"), e("4"), e("2"))
	endpoint2 := hash.Get("11")
	assert.Equal(t, "2", endpoint2.Key())
	endpoint4 := hash.Get("33")
	assert.Equal(t, "4", endpoint4.Key())
	assert.Equal(t, 2, hash.totalLoad)

	// Remove endpoints that had requests pending. The load should be
	// adjustedAnd and Put() be a no op.
	hash.RemoveEndpoints(endpoint2, endpoint4)
	assert.Equal(t, 0, hash.totalLoad)
	hash.Put(endpoint2)
	hash.Put(endpoint4)
}

func TestLoadOK(t *testing.T) {
	tests := []struct {
		totalLoad, numEndpoints, endpointLoad int
		factor                                float64
		expected                              bool
	}{
		{99, 1, 99, 1.20, true},
		{99, 4, 23, 1.20, true},
		{99, 4, 29, 1.20, true},
		{99, 4, 30, 1.20, false},
	}

	for _, test := range tests {
		got := loadOK(test.totalLoad, test.numEndpoints, test.endpointLoad, test.factor)
		assert.Equal(t, test.expected, got)
	}
}

type testEndpoint struct {
	key  string
	load int
}

type boundedState []testEndpoint

func (s boundedState) totalLoad() int {
	load := 0
	for i := range s {
		load += s[i].load
	}
	return load
}

func makeTestHash(loadFactor float64, state boundedState) *Consistent {
	hash := NewConsistent(ConsistentConfig{
		ReplicationCount: 3,
		Hash:             testHash,
		LoadFactor:       loadFactor,
	})

	for _, endpoint := range state {
		hash.AddEndpoints(e(endpoint.key))
		info := hash.info(endpoint.key)
		info.load = endpoint.load
	}

	hash.totalLoad = state.totalLoad()

	return hash
}

type getOperation struct {
	key              string
	expectedEndpoint string
	expectedLoad     int
}

func TestBoundedLoad(t *testing.T) {
	tests := []struct {
		loadFactor float64
		state      boundedState
		ops        []getOperation
	}{
		{
			1.20,
			boundedState{
				{"6", 30},
				{"4", 22},
				{"2", 24},
				{"7", 23},
			},
			[]getOperation{
				// Node 6 is already at the max allowed load, request should be handled by
				// node 7.
				{"15", "7", 24},
				// Node 4 not too loaded, so should handle that Get()
				{"13", "4", 23},
			},
		},
		// 1.001 * 30 < 31, make sure we can fit the request on the first node hit.
		{
			1.001,
			boundedState{
				{"6", 30},
				{"4", 30},
				{"2", 30},
				{"7", 30},
			},
			[]getOperation{
				// Node 4 not too loaded, so should handle that Get()
				{"13", "4", 31},
			},
		},
	}

	for _, test := range tests {
		hash := makeTestHash(test.loadFactor, test.state)
		for _, op := range test.ops {
			// Get
			endpoint := hash.Get(op.key)
			assert.Equal(t, op.expectedEndpoint, endpoint.Key())
			assert.Equal(t, op.expectedLoad, hash.info(endpoint.Key()).load)
			assert.Equal(t, hash.totalLoad, test.state.totalLoad()+1)

			// Put
			hash.Put(endpoint)
			assert.Equal(t, hash.totalLoad, test.state.totalLoad())
		}

	}
}

func BenchmarkGet8(b *testing.B)   { benchmarkGet(b, 8) }
func BenchmarkGet32(b *testing.B)  { benchmarkGet(b, 32) }
func BenchmarkGet128(b *testing.B) { benchmarkGet(b, 128) }
func BenchmarkGet512(b *testing.B) { benchmarkGet(b, 512) }

func benchmarkGet(b *testing.B, shards int) {

	hash := NewConsistent(ConsistentConfig{ReplicationCount: 50})

	var buckets []Endpoint
	for i := 0; i < shards; i++ {
		buckets = append(buckets, e(fmt.Sprintf("shard-%d", i)))
	}

	hash.AddEndpoints(buckets...)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		hash.Get(buckets[i&(shards-1)].Key())
	}
}
