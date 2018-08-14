package main

import (
	"flag"
	"net"
	"os"
	"testing"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/hashicorp/consul/api"
	echo "github.com/nicholasjackson/grpc-consul-resolver/functional_tests/grpc"
	"google.golang.org/grpc"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

const consulAddr = "http://localhost:8500"

var consulClient *api.Client
var echoClient echo.EchoServiceClient
var preparedQueryID string

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
	s.Step(`^I have a prepared query$`, iHaveAPreparedQuery)
	s.Step(`^(\d+) services are started$`, nServicesAreRunningAndRegistered)
	s.Step(`^(\d+) services are removed$`, nServicesAreStopped)
	s.Step(`^I call use the client (\d+) times$`, iCallUseTheClientTimes)
	s.Step(`^I call the client (\d+) times with a query$`, iCallTheClientTimesWithAQuery)
	s.Step(`^I expect (\d+) different endpoints to have been called$`, iExpectDifferentEndpointsToHaveBeenCalled)
}
