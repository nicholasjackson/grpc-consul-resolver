package resolver

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatesNewResolver(t *testing.T) {
	r := NewResolver(&MockConsulHealth{})

	assert.NotNil(t, r)
}

func TestResolveRerturnsWatcher(t *testing.T) {
	r := NewResolver(&MockConsulHealth{})

	w, err := r.Resolve("target")

	assert.NoError(t, err)
	assert.NotNil(t, w)
}
