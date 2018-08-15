package resolver

import (
	"testing"

	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	"github.com/stretchr/testify/assert"
)

func TestCreatesNewResolver(t *testing.T) {
	r := NewResolver(&catalog.MockQuery{})

	assert.NotNil(t, r)
}

func TestResolveRerturnsWatcher(t *testing.T) {
	r := NewResolver(&catalog.MockQuery{})

	w, err := r.Resolve("target")

	assert.NoError(t, err)
	assert.NotNil(t, w)
}

func TestReverseLookupReturnsEntry(t *testing.T) {
	t.Fatal("Pending")
}
