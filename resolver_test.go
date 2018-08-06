package resolver

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreatesNewResolver(t *testing.T) {
	r := NewResolver(1*time.Millisecond, &MockConsulHealth{})

	assert.NotNil(t, r)
}

func TestResolveRerturnsWatcher(t *testing.T) {
	r := NewResolver(1*time.Millisecond, &MockConsulHealth{})

	w, err := r.Resolve("target")

	assert.NoError(t, err)
	assert.NotNil(t, w)
}
