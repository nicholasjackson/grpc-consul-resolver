package resolver

import (
	"context"
	"testing"

	"github.com/hashicorp/consul/agent/connect"
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

func TestStaticResolverReturnsStaticResolver(t *testing.T) {
	r := NewResolver(&catalog.MockQuery{})

	w, _ := r.Resolve("target")
	w.(*ConsulWatcher).addressCache["localhost:8080"] = catalog.ServiceEntry{
		Addr: "localhost:8181",
		CertURI: &connect.SpiffeIDService{
			Host:       "abc123",
			Namespace:  "default",
			Datacenter: "dc1",
			Service:    "tester",
		},
	}

	sr, _ := r.StaticResolver("localhost:8080")
	addr, certURI, err := sr.Resolve(context.Background())

	assert.NoError(t, err)
	assert.Equal(t, "localhost:8181", addr)
	assert.Equal(t, "spiffe://abc123/ns/default/dc/dc1/svc/tester", certURI.URI().String())
}
