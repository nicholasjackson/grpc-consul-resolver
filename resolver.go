package resolver

import (
	"time"

	"google.golang.org/grpc/naming"
)

// ConsulResolver is a service resolver for gRPC load balancing
// example usage:
// r := resolver.NewResolver(10*time.Second, consulClient.Health())
// lb := grpc.RoundRobin(r)
//
// c, err := grpc.Dial("test_grpc", grpc.WithInsecure(), grpc.WithBalancer(lb))
type ConsulResolver struct {
	client       ConsulHealth
	PollInterval time.Duration
}

// NewResolver returns a new ConsulResolver with the given client
// PollInterval is set to a sensible default of 60 seconds
func NewResolver(c ConsulHealth) *ConsulResolver {
	return &ConsulResolver{client: c, PollInterval: 60 * time.Second}
}

// Resolve called internally by the load balancer
func (g *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return NewConsulWatcher(
		target,
		g.client,
		g.PollInterval,
	), nil
}
