package resolver

import (
	"google.golang.org/grpc/naming"
)

// ConsulResolver is a service resolver for gRPC load balancing
type ConsulResolver struct {
	client ConsulHealth
}

// NewResolver returns a new ConsulResolver with the given client
func NewResolver(c ConsulHealth) *ConsulResolver {
	return &ConsulResolver{c}
}

// Resolve called internally by the load balancer
func (g *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return &ConsulWatcher{}, nil
}
