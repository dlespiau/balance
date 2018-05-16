// Copyright 2013 Google Inc.
// Copyright 2018 Weaveworks
//
// Originally from:
//  https://github.com/golang/groupcache/blob/master/consistenthash/consistenthash_test.go

package algo

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type e string

func (e e) Key() string { return string(e) }

func TestHashing(t *testing.T) {

	// Override the hash function to return easier to reason about values. Assumes
	// the keys can be converted to an integer.
	hash := NewConsistent(ConsistentConfig{
		ReplicationCount: 3,
		Hash: func(key []byte) uint32 {
			i, err := strconv.Atoi(string(key))
			if err != nil {
				panic(err)
			}
			return uint32(i)
		},
	})

	// Given the above hash function, this will give replicas with "hashes":
	// 2, 4, 6, 12, 14, 16, 22, 24, 26
	hash.AddEndpoints(e("6"), e("4"), e("2"))
	assert.Equal(t, []int{2, 4, 6, 12, 14, 16, 22, 24, 26}, hash.keys)

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

	// 27 should now map to 8.
	testCases["27"] = e("8")

	for k, v := range testCases {
		assert.Equal(t, v, hash.Get(k))
	}

	// Removes 8, 18, 28
	hash.RemoveEndpoints(e("8"))
	assert.Equal(t, []int{2, 4, 6, 12, 14, 16, 22, 24, 26}, hash.keys)

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
