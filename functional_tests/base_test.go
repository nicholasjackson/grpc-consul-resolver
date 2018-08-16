package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	resolver "github.com/nicholasjackson/grpc-consul-resolver"
	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	echo "github.com/nicholasjackson/grpc-consul-resolver/functional_tests/grpc"
	"google.golang.org/grpc"
)

func thatConsulIsRunning() error {
	l, err := consulClient.Status().Leader()
	if err != nil {
		return err
	}

	if l == "" {
		return fmt.Errorf("No consul leader")
	}

	return nil
}

func nServicesAreRunningAndRegistered(arg1 int) error {
	for i := 0; i < arg1; i++ {
		port := rand.Intn(1000) + 8000

		err := runGRPCServer(port)
		if err != nil {
			return err
		}
	}

	return nil
}

func nServicesAreStopped(arg1 int) error {
	stopped := 0

	for _, v := range gRPCServers {
		if stopped < arg1 {
			stopGRPCServer(v)
			stopped++
		}
	}

	return nil
}

func iCallUseTheClientTimes(arg1 int) error {
	err := initServiceClientIfNeeded()
	if err != nil {
		return err
	}

	return callClient(echoClient, arg1)
}

func iCallTheClientTimesWithAQuery(arg1 int) error {
	err := initQueryClientIfNeeded()
	if err != nil {
		return err
	}

	return callClient(echoClient, arg1)
}

func iCallTheConnectEnabledClientTimes(arg1 int) error {
	err := initConnectServiceClientIfNeeded()
	if err != nil {
		return err
	}

	return callClient(echoClient, arg1)
}

func callClient(c echo.EchoServiceClient, times int) error {
	for i := 0; i < times; i++ {
		msg, err := c.Echo(context.Background(), &echo.Message{Data: "hello world"})
		if err != nil {
			return err
		}

		responses = append(responses, msg.Data)

		time.Sleep(1 * time.Second)
	}

	return nil
}

func iExpectDifferentEndpointsToHaveBeenCalled(arg1 int) error {
	// count the unique responses
	uniques := make([]string, 0)

	for _, r := range responses {
		captured := false
		for _, u := range uniques {
			if r == u {
				captured = true
			}
		}

		if !captured {
			uniques = append(uniques, r)
		}
	}

	if len(uniques) != arg1 {
		return fmt.Errorf("Expecting %d unique responses, got %d", arg1, len(uniques))
	}

	return nil
}

func iHaveAPreparedQuery() error {
	pqo := &api.PreparedQueryDefinition{
		Name: "prepared_query",
		Service: api.ServiceQuery{
			Service:     "test_grpc",
			OnlyPassing: true,
		},
	}

	var err error
	preparedQueryID, _, err = consulClient.PreparedQuery().Create(pqo, nil)

	return err
}

func setup(i interface{}) {
	responses = make([]string, 0)
	gRPCServers = make(map[string]*gRPCServer)

	conf := api.DefaultConfig()
	conf.Address = consulAddr
	consulClient, _ = api.NewClient(conf)
}

func cleanup(i interface{}, err error) {
	// stop the gRPC servers
	for _, s := range gRPCServers {
		stopGRPCServer(s)
	}

	echoClient = nil

	// remove any prepared queries
	if preparedQueryID != "" {
		consulClient.PreparedQuery().Delete(preparedQueryID, nil)
		preparedQueryID = ""
	}
}

func runGRPCServer(port int) error {
	addr := fmt.Sprintf("localhost:%d", port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	echo.RegisterEchoServiceServer(s, &echo.EchoServiceServerImpl{ID: addr})
	go s.Serve(lis)

	// register with Consul
	err = consulClient.Agent().ServiceRegister(
		&api.AgentServiceRegistration{
			ID:      addr,
			Name:    "test_grpc",
			Port:    port,
			Address: "localhost",
			Connect: &api.AgentServiceConnect{
				Proxy: &api.AgentServiceConnectProxy{
					Config: map[string]interface{}{"bind_port": port + 1000},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	gRPCServers[addr] = &gRPCServer{
		address: addr,
		server:  s,
		socket:  lis,
	}

	return nil
}

func stopGRPCServer(s *gRPCServer) {
	consulClient.Agent().ServiceDeregister(s.address)
	s.server.GracefulStop()
	s.socket.Close()

	delete(gRPCServers, s.address)
}

func initServiceClientIfNeeded() error {
	if echoClient != nil {
		return nil
	}

	sq := catalog.NewServiceQuery(consulClient, false)
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests

	lb := grpc.RoundRobin(r)

	// create a new client and wait to establish a connection before returnig
	c, err := grpc.Dial(
		"test_grpc",
		grpc.WithInsecure(),
		grpc.WithBalancer(lb),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)

	if err != nil {
		return fmt.Errorf("Setup error creating grpc client %s", err.Error())
	}

	echoClient = echo.NewEchoServiceClient(c)

	return nil
}

func initConnectServiceClientIfNeeded() error {
	if echoClient != nil {
		return nil
	}

	fmt.Println("init client")

	service, err := connect.NewService("test_grpc", consulClient)
	if err != nil {
		fmt.Println("Unable to create service", "error", err)
		return err
	}

	sq := catalog.NewServiceQuery(consulClient, true)
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests

	lb := grpc.RoundRobin(r)

	// create a new client and wait to establish a connection before returnig
	c, err := grpc.Dial(
		"test_grpc",
		grpc.WithInsecure(),
		grpc.WithBalancer(lb),
		grpc.WithBlock(),
		grpc.WithTimeout(10*time.Second),
		grpc.WithDialer(func(addr string, t time.Duration) (net.Conn, error) {
			sr, err := r.StaticResolver(addr)
			if err != nil {
				return nil, err
			}

			return service.Dial(context.Background(), sr)
		}),
	)

	if err != nil {
		return fmt.Errorf("Setup error creating grpc client %s", err.Error())
	}

	echoClient = echo.NewEchoServiceClient(c)

	return nil
}

func initQueryClientIfNeeded() error {
	if echoClient != nil {
		return nil
	}

	sq := catalog.NewPreparedQuery(consulClient.PreparedQuery())
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests

	lb := grpc.RoundRobin(r)

	// create a new client and wait to establish a connection before returnig
	c, err := grpc.Dial(
		"prepared_query",
		grpc.WithInsecure(),
		grpc.WithBalancer(lb),
		grpc.WithBlock(),
		grpc.WithTimeout(5*time.Second),
	)

	if err != nil {
		return fmt.Errorf("Setup error creating grpc client %s", err.Error())
	}

	echoClient = echo.NewEchoServiceClient(c)

	return nil
}
