package main

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"os"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/hashicorp/consul/api"
	resolver "github.com/nicholasjackson/grpc-consul-resolver"
	"github.com/nicholasjackson/grpc-consul-resolver/functional_tests/grpc"
	"google.golang.org/grpc"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

const consulAddr = "http://localhost:8500"

var consulClient *api.Client
var responses []string

type gRPCServer struct {
	address string
	server  *grpc.Server
	socket  net.Listener
}

var gRPCServers map[string]*gRPCServer

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
}

func TestMain(m *testing.M) {
	flag.Parse()
	opt.Paths = flag.Args()

	status := godog.RunWithOptions("godogs", func(s *godog.Suite) {
		FeatureContext(s)
	}, opt)

	if st := m.Run(); st > status {
		status = st
	}
	os.Exit(status)
}

func FeatureContext(s *godog.Suite) {
	s.BeforeScenario(setup)
	s.AfterScenario(cleanup)

	s.Step(`^that Consul is running$`, thatConsulIsRunning)
	s.Step(`^(\d+) services are started$`, nServicesAreRunningAndRegistered)
	s.Step(`^(\d+) services are removed$`, nServicesAreStopped)
	s.Step(`^I call use the client (\d+) times$`, iCallUseTheClientTimes)
	s.Step(`^I expect (\d+) different endpoints to have been called$`, iExpectDifferentEndpointsToHaveBeenCalled)
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
}

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
	r := resolver.NewResolver(consulClient.Health())
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
		return fmt.Errorf("Error creating grpc client %s", err.Error())
	}

	defer c.Close()

	cc := echo.NewEchoServiceClient(c)

	for i := 0; i < arg1; i++ {
		msg, err := cc.Echo(context.Background(), &echo.Message{Data: "hello world"})
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

func stopGRPCServer(s *gRPCServer) {
	consulClient.Agent().ServiceDeregister(s.address)
	s.server.GracefulStop()
	s.socket.Close()

	delete(gRPCServers, s.address)
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
	err = consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      addr,
		Name:    "test_grpc",
		Port:    port,
		Address: "localhost",
	})
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
