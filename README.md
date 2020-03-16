# gRPC Resolver for Consul

[![CircleCI](https://circleci.com/gh/nicholasjackson/grpc-consul-resolver.svg?style=svg)](https://circleci.com/gh/nicholasjackson/grpc-consul-resolver)
[![GoDoc](https://godoc.org/github.com/nicholasjackson/grpc-consul-resolver?status.svg)](https://godoc.org/github.com/nicholasjackson/grpc-consul-resolver)

This repository implements a naming.Resolver for Consul which can be used with gRPC load balancers.

For information on load balancing concepts with gRPC please see the documentation:   
[https://github.com/grpc/grpc/blob/master/doc/load-balancing.md](https://github.com/grpc/grpc/blob/master/doc/load-balancing.md)

When creating a gRPC load balancer a resolver must be passed as a dependency:

```
func RoundRobin(r naming.Resolver) Balancer
```

It is the resolver's job is to determine the endpoints for the given service name.  When `grpc.Dial` has been setup with a load balancer and you make a call to a service, internal the gRPC framework
requests an endpoint from the load balancer.  The load balancer gets this information from the resolver at creation time, this is supplied by the resolver function `Next()`.   

Once this first batch of endpoints has been retrieved it is the resolvers job to watch the service catalog and to return any updated information, letting the load balancer know of any added or deleted records.  The `Next()` function blocks until there is updated service information, internally inside the load balancer the resolvers `Next()` function is continually called, a return from this function informs it that it needs to update the internal endpoint list.  

![](https://github.com/grpc/grpc/raw/master/doc/images/load-balancing.png)

Connections in gRPC are persistent, it is common to have a single client which is shared across all go routines, it is the clients job to marshal access to the connections, spawning additional connections as required.  When a load balancer is used then the client will maintain at least one connection for each endpoint in the load balanced list.  With each call to a service these will be rotated according to the policy implemented by the load balancer.  For example if you have two endpoints `127.0.0.1:8080` and `127.0.0.1:8081` using the built in `RoundRobin` load balancer would ensure that every call to a service endpoint would rotate through the endpoints returned from the resolver in turn.

Internally this implementation of a gRPC Resolver leverages Consuls Service Catalog, endpoints are retrieved from the catalog based
on their registered name.  The resolver continually polls Consul (60 seconds by default) to ensure the endpoint list is kept up to date.

## Basic usage:
```
serviceName := "myService"
r := resolver.NewServiceQueryResolver("http://consulAddr:8500", serviceName)

// use the default poll interval of 60 seconds
// the poll interval can be changed by setting the resolvers PollInterval field
// r.PollInterval = 10 * time.Second

// Create the gRPC load balancer
lb := grpc.RoundRobin(r)

// create a new gRPC client connection
c, err := grpc.Dial(
	serviceName,
	grpc.WithInsecure(),
	grpc.WithBalancer(lb),
	grpc.WithTimeout(5*time.Second),
)

// create the instance of our test client
cc := echo.NewEchoServiceClient(c)

// call the service method
// The first call would be routed to the first endpoint which was returned from 
// the Resolver and thus the Consul Service Catalog
cc.Echo(context.Background(), &echo.Message{Data: "hello world"})

// call the service method again
// The second call would create a new connection, this time the endpoint used 
// would be the second listed in the Consul Service Catalog
cc.Echo(context.Background(), &echo.Message{Data: "hello world"})
```

## Consul Connect usage:
```
r, dialer, _ := resolver.NewConnectServiceQueryResolver("http://consulAddr:8500","my_service")
lb := grpc.RoundRobin(r)

// create a new gRPC client connection
c, err := grpc.Dial(
	"test_grpc",
	grpc.WithInsecure(),
	grpc.WithBalancer(lb),
	grpc.WithTimeout(5*time.Second),
  dialer,
)

```

## Testing
This package has both `unit` and `integration` tests, the unit tests are pure Go tests with mocks replacing the dependency for Consul.  To execute unit tests:

```bash
$ make test_unit

# or

$ go test -v -race .

=== RUN   TestCreatesNewResolver
--- PASS: TestCreatesNewResolver (0.00s)
=== RUN   TestResolveRerturnsWatcher
--- PASS: TestResolveRerturnsWatcher (0.00s)
=== RUN   TestNewConsulWatcherReturnsWatcher
--- PASS: TestNewConsulWatcherReturnsWatcher (0.00s)
=== RUN   TestNextReturnsErrorWhenConsulError
--- PASS: TestNextReturnsErrorWhenConsulError (0.00s)
=== RUN   TestNextReturnsInitialUpdatesFromConsul
--- PASS: TestNextReturnsInitialUpdatesFromConsul (0.00s)
=== RUN   TestNextReturnsInitialUpdatesFromConsulSetsNodeWhenNoAddr
--- PASS: TestNextReturnsInitialUpdatesFromConsulSetsNodeWhenNoAddr (0.00s)
=== RUN   TestNextReturnsUpdatesContainingAddedItemsFromConsul
--- PASS: TestNextReturnsUpdatesContainingAddedItemsFromConsul (0.00s)
=== RUN   TestNextReturnsUpdatesContainingDeletedItemsFromConsul
--- PASS: TestNextReturnsUpdatesContainingDeletedItemsFromConsul (0.00s)
=== RUN   TestNextBlocksWhenNoChangesFromConsul
--- PASS: TestNextBlocksWhenNoChangesFromConsul (0.05s)
PASS
ok      github.com/nicholasjackson/grpc-consul-resolver 1.073s
```

In addition to the unit tests there is also an integration test suite, the test suite requires a `Consul` server to be running on localhost with the default ports as a dependency. The integration tests start two dummy gRPC servers and register them with the Consul server's Service Catalog.  A gRPC client is then created to ensure the function of the Resolver.  Integration tests can be found in the sub folder `./functional_tests`, the [GoDog](https://github.com/DATA-DOG/godog) Cucumber BDD framework is used to execute these tests.  To execute theintegration tests:

```bash
$ make test_functional

# or

$ cd functional_tests
$ go test -v --godog.format=pretty --godog.random

Feature: As a developer, I want to ensure that the
  loabalancer functions correctly

  Scenario: Calls two different upstreams                   # features/consul_service.feature:10
    Given that Consul is running                            # main_test.go:47 -> thatConsulIsRunning
    And the services are running and registered             # main_test.go:60 -> theServicesAreRunningAndRegistered
Server id localhost:7711 Echo request hello world
Server id localhost:7712 Echo request hello world
    When I call use the client 2 times                      # main_test.go:88 -> iCallUseTheClientTimes
    Then I expect 2 different endpoints to have been called # main_test.go:123 -> iExpectDifferentEndpointsToHaveBeenCalled

  Scenario: Calls one upstream                              # features/consul_service.feature:4
    Given that Consul is running                            # main_test.go:47 -> thatConsulIsRunning
    And the services are running and registered             # main_test.go:60 -> theServicesAreRunningAndRegistered
Server id localhost:7712 Echo request hello world
    When I call use the client 1 times                      # main_test.go:88 -> iCallUseTheClientTimes
    Then I expect 1 different endpoints to have been called # main_test.go:123 -> iExpectDifferentEndpointsToHaveBeenCalled

2 scenarios (2 passed)
8 steps (8 passed)
3.664543687s

Randomized with seed: 54915
testing: warning: no tests to run
PASS
ok      github.com/nicholasjackson/grpc-consul-resolver/functional_tests        3.699s
``` 



## TODO
[x] Implement Consul Connect Services lookup   
[x] Implement prepared queries  
[ ] Implement prepared queries with Connect Services  
[ ] Finish implementing query options  
[ ] Investigate why functional tests hang on CircleCI but run fine locally  
[ ] Tidy readme and documentation  
