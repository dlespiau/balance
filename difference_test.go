package balance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDiff(t *testing.T) {
	tests := []struct {
		a, b     []Endpoint
		expected []Endpoint
	}{
		{el("a", "b"), el("b"), el("a")},
		{el("a", "b"), el("c"), el("a", "b")},
		{el("a", "b"), el("a", "b"), nil},
	}

	for _, test := range tests {
		assert.Equal(t, keys(test.expected), keys(diff(test.a, test.b)))
	}
}

func chunkOperations(chunks []chunk) []differenceOperation {
	ops := make([]differenceOperation, 0, len(chunks))
	for _, chunk := range chunks {
		ops = append(ops, chunk.operation)
	}
	return ops
}

func chunkKeys(chunks []chunk) []string {
	keys := make([]string, len(chunks))
	for _, chunk := range chunks {
		keys = append(keys, chunk.endpoint.Key())
	}
	return keys

}

func assertChunks(t *testing.T, expected, got []chunk) {
	assert.Equal(t, len(expected), len(got))
	assert.Equal(t, chunkOperations(expected), chunkOperations(got))
	assert.Equal(t, chunkKeys(expected), chunkKeys(got))
}

func TestDifference(t *testing.T) {
	tests := []struct {
		a, b     []Endpoint
		expected []chunk
	}{
		{
			[]Endpoint{},
			[]Endpoint{e("b"), e("a")},
			[]chunk{
				{add, e("b")},
				{add, e("a")},
			},
		},
		{
			[]Endpoint{e("b"), e("a")},
			[]Endpoint{},
			[]chunk{
				{del, e("b")},
				{del, e("a")},
			},
		},
		{
			[]Endpoint{e("a")},
			[]Endpoint{e("b"), e("a")},
			[]chunk{
				{add, e("b")},
			},
		},
		{
			[]Endpoint{e("a"), e("b")},
			[]Endpoint{e("a")},
			[]chunk{
				{del, e("b")},
			},
		},
		{
			[]Endpoint{e("b"), e("d"), e("c")},
			[]Endpoint{e("d"), e("e"), e("f")},
			[]chunk{
				{del, e("b")},
				{del, e("c")},
				{add, e("e")},
				{add, e("f")},
			},
		},
	}

	for _, test := range tests {
		assertChunks(t, test.expected, difference(test.a, test.b))
	}
}
