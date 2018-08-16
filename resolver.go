package resolver

import (
	"fmt"
	"time"

	"github.com/hashicorp/consul/connect"
	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	"google.golang.org/grpc/naming"
)

// ConsulResolver is a service resolver for gRPC load balancing
// example usage:
// r := resolver.NewResolver(10*time.Second, consulClient.Health())
// lb := grpc.RoundRobin(r)
//
// c, err := grpc.Dial("test_grpc", grpc.WithInsecure(), grpc.WithBalancer(lb))
type ConsulResolver struct {
	query        catalog.Query
	PollInterval time.Duration
	watchers     map[string]*ConsulWatcher
}

// NewResolver returns a new ConsulResolver with the given client
// PollInterval is set to a sensible default of 60 seconds
func NewResolver(q catalog.Query) *ConsulResolver {
	return &ConsulResolver{query: q, PollInterval: 60 * time.Second, watchers: make(map[string]*ConsulWatcher)}
}

// Resolve called internally by the load balancer
func (g *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	w := NewConsulWatcher(
		target,
		g.query,
		g.PollInterval,
	)

	g.watchers[target] = w

	return w, nil
}

// StaticResolver allows fetching the service entry from the cache
// this is a required function for the Connect static resolver which needs details from the ServiceEntry
func (g *ConsulResolver) StaticResolver(address string) (*connect.StaticResolver, error) {
	// find the details in the cache
	for _, v := range g.watchers {
		se, ok := v.addressCache[address]

		if ok {
			return &connect.StaticResolver{
				Addr:    se.Addr,
				CertURI: se.CertURI,
			}, nil
		}
	}

	return nil, fmt.Errorf("Unable to resolve address")
}
