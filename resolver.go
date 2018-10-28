package resolver

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	"google.golang.org/grpc"
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

// NewServiceQueryResolver is a convenience constructor which returns a resolver for the given consul server
func NewServiceQueryResolver(consulAddr string) *ConsulResolver {
	conf := api.DefaultConfig()
	conf.Address = consulAddr
	consulClient, _ := api.NewClient(conf)

	sq := catalog.NewServiceQuery(consulClient, false)

	return NewResolver(sq)
}

// NewConnectServiceQueryResolver is a convenience constructor which returns a consul connect enabled resolver for the given consul server
func NewConnectServiceQueryResolver(consulAddr, serviceName string) (*ConsulResolver, grpc.DialOption, error) {
	conf := api.DefaultConfig()
	conf.Address = consulAddr
	consulClient, _ := api.NewClient(conf)

	connectService, err := connect.NewService(serviceName, consulClient)
	if err != nil {
		return nil, nil, fmt.Errorf("Unable to create connect service %s", err)
	}

	sq := catalog.NewServiceQuery(consulClient, true)
	r := NewResolver(sq)

	// We need to create a custom dialer for gRPC, instead of using the built in
	// net.Dial we will use the Dial method from the Consul Connect service.
	// This ensures that mTLS secures the transport and the upstream service
	// identity is valid
	withDialer := grpc.WithDialer(func(addr string, t time.Duration) (net.Conn, error) {
		// Dial in the Connect package requires a service resolver which returns
		// the upstream address and the certificate info retrieved from consul
		// when the service catalog was queried.
		// Because service resolution has allready been carried out by the gRPC
		// loadbalancer through the Resolver we can use the reverse lookup which
		// takes an endpoint address as a parameter to return a connect StaticResolver
		// containing the information required for the connection.
		sr, err := r.StaticResolver(addr)
		if err != nil {
			return nil, err
		}

		return connectService.Dial(context.Background(), sr)
	})

	return r, withDialer, nil
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
