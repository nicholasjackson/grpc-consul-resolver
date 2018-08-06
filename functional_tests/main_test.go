package main

import (
	"context"
	"flag"
	"fmt"
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

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)
	responses = make([]string, 0)
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

func theServicesAreRunningAndRegistered() error {
	go runGRPCServer("localhost:7711")
	go runGRPCServer("localhost:7712")

	err := consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      "test_grpc1",
		Name:    "test_grpc",
		Port:    7711,
		Address: "localhost",
	})
	if err != nil {
		return err
	}

	err = consulClient.Agent().ServiceRegister(&api.AgentServiceRegistration{
		ID:      "test_grpc2",
		Name:    "test_grpc",
		Port:    7712,
		Address: "localhost",
	})
	if err != nil {
		return err
	}

	return nil
}

func iCallUseTheClientTimes(arg1 int) error {
	lb := grpc.RoundRobin(
		resolver.NewResolver(1*time.Second, consulClient.Health()),
	)

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

func FeatureContext(s *godog.Suite) {
	s.BeforeScenario(setup)
	s.AfterScenario(cleanup)

	s.Step(`^that Consul is running$`, thatConsulIsRunning)
	s.Step(`^the services are running and registered$`, theServicesAreRunningAndRegistered)
	s.Step(`^I call use the client (\d+) times$`, iCallUseTheClientTimes)
	s.Step(`^I expect (\d+) different endpoints to have been called$`, iExpectDifferentEndpointsToHaveBeenCalled)
}

func setup(i interface{}) {
	conf := api.DefaultConfig()
	conf.Address = consulAddr
	consulClient, _ = api.NewClient(conf)
}

func cleanup(i interface{}, err error) {
	consulClient.Agent().ServiceDeregister("test_grpc1")
	consulClient.Agent().ServiceDeregister("test_grpc2")
}

func runGRPCServer(listen string) {
	lis, err := net.Listen("tcp", listen)
	if err != nil {
		fmt.Println(err)
	}

	grpcServer := grpc.NewServer()
	echo.RegisterEchoServiceServer(grpcServer, &echo.EchoServiceServerImpl{ID: listen})
	grpcServer.Serve(lis)
}
