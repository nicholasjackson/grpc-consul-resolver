# GRPC Consul Resolver


Basic usage:
```
// Create a consul client
conf := api.DefaultConfig()
conf.Address = "http://localhost:8500"
consulClient, _ = api.NewClient(conf)

// create an instance of the load balancer with our resolver
lb := grpc.RoundRobin(
	resolver.NewResolver(10*time.Second, consulClient.Health()),
)

// create a new client and wait to establish a connection before returning
c, err := grpc.Dial(
	"test_grpc",
	grpc.WithInsecure(),
	grpc.WithBalancer(lb),
	grpc.WithBlock(),
	grpc.WithTimeout(5*time.Second),
)
```


