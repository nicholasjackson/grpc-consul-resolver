package resolver

import (
	"time"

	"google.golang.org/grpc/naming"
)

// ConsulResolver is a service resolver for gRPC load balancing
type ConsulResolver struct {
	client        ConsulHealth
	watchDuration time.Duration
}

// NewResolver returns a new ConsulResolver with the given client
func NewResolver(watchDuration time.Duration, c ConsulHealth) *ConsulResolver {
	return &ConsulResolver{client: c, watchDuration: watchDuration}
}

// Resolve called internally by the load balancer
func (g *ConsulResolver) Resolve(target string) (naming.Watcher, error) {
	return NewConsulWatcher(
		target,
		g.client,
		g.watchDuration,
	), nil
}
