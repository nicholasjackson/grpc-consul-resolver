package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/godog"
	"github.com/DATA-DOG/godog/colors"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/connect"
	resolver "github.com/nicholasjackson/grpc-consul-resolver"
	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	echo "github.com/nicholasjackson/grpc-consul-resolver/functional_tests/grpc"
	"google.golang.org/grpc"
)

var opt = godog.Options{Output: colors.Colored(os.Stdout)}

var consulAddr = "http://localhost:8500"
var serviceBind = "localhost"
var serviceName = "test_grpc"
var preparedQueryName = "test_grpc_query"

var consulClient *api.Client
var grpcClient *grpc.ClientConn
var connectService *connect.Service
var echoClient echo.EchoServiceClient
var preparedQueryID string

var responses []string

type gRPCServer struct {
	id           string
	address      string
	server       *grpc.Server
	socket       net.Listener
	proxyCommand *exec.Cmd
}

type proxies map[string]string

var gRPCServers map[string]*gRPCServer

func init() {
	godog.BindFlags("godog.", flag.CommandLine, &opt)

	envAddr := os.Getenv("CONSUL_HTTP_ADDR")
	if envAddr != "" {
		consulAddr = envAddr
	}

	envAddr = os.Getenv("BIND_ADDR")
	if envAddr != "" {
		serviceBind = envAddr
	}
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
	s.Step(`^I call the connect enabled client (\d+) times$`, iCallTheConnectEnabledClientTimes)
	s.Step(`^I call the client (\d+) times with a query$`, iCallTheClientTimesWithAQuery)
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

	grpcClient.Close()
	echoClient = nil

	// remove any prepared queries
	if preparedQueryID != "" {
		consulClient.PreparedQuery().Delete(preparedQueryID, nil)
		preparedQueryID = ""
	}

	if connectService != nil {
		connectService.Close()
	}
}

func runGRPCServer(port int) error {
	addr := fmt.Sprintf("%s:%d", serviceBind, port)

	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	s := grpc.NewServer()
	echo.RegisterEchoServiceServer(s, &echo.EchoServiceServerImpl{ID: addr})
	go s.Serve(lis)

	// register with Consul
	id := strings.Replace(addr, ":", "-", -1)
	err = consulClient.Agent().ServiceRegister(
		&api.AgentServiceRegistration{
			ID:      id,
			Name:    serviceName,
			Port:    port,
			Address: serviceBind,
			Check: &api.AgentServiceCheck{
				CheckID:                        id,
				DeregisterCriticalServiceAfter: "1m",
				TCP:                            addr,
				Interval:                       "1s",
			},
		},
	)
	if err != nil {
		return err
	}

	// start the proxy
	cmd := startProxy(serviceName, serviceBind, port)

	gRPCServers[addr] = &gRPCServer{
		id:           id,
		address:      addr,
		server:       s,
		socket:       lis,
		proxyCommand: cmd,
	}

	return nil
}

func startProxy(serviceName, serviceAddress string, servicePort int) *exec.Cmd {
	proxyPort := servicePort + 1000

	cmd := exec.Command(
		"consul",
		"connect",
		"proxy",
		"-service", serviceName,
		"-http-addr", consulAddr,
		"-listen", fmt.Sprintf("%s:%d", serviceAddress, proxyPort),
		"-service-addr", fmt.Sprintf("%s:%d", serviceAddress, servicePort),
		"-register",
		"-register-id", fmt.Sprintf("%s-%d", serviceName, servicePort),
	)

	err := cmd.Start()
	if err != nil {
		log.Println(err)
	}
	return cmd
}

func stopGRPCServer(s *gRPCServer) {
	consulClient.Agent().ServiceDeregister(s.id)
	s.server.GracefulStop()
	s.socket.Close()

	// stop the proxy
	s.proxyCommand.Process.Signal(os.Interrupt)

	delete(gRPCServers, s.address)
}

func initServiceClientIfNeeded() error {
	sq := catalog.NewServiceQuery(consulClient, false)
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests
	lb := grpc.RoundRobin(r)

	return initClient(serviceName, grpc.WithBalancer(lb))
}

func initConnectServiceClientIfNeeded() error {
	var err error
	connectService, err = connect.NewService(serviceName, consulClient)
	if err != nil {
		return fmt.Errorf("Unable to create connect service %s", err)
	}

	sq := catalog.NewServiceQuery(consulClient, true)
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests

	dialer := grpc.WithDialer(func(addr string, t time.Duration) (net.Conn, error) {
		sr, err := r.StaticResolver(addr)
		if err != nil {
			return nil, err
		}

		return connectService.Dial(context.Background(), sr)
	})

	lb := grpc.RoundRobin(r)

	return initClient(serviceName, grpc.WithBalancer(lb), dialer)
}

func initQueryClientIfNeeded() error {
	sq := catalog.NewPreparedQuery(consulClient.PreparedQuery())
	r := resolver.NewResolver(sq)
	r.PollInterval = 1 * time.Second // override poll interval for tests
	lb := grpc.RoundRobin(r)

	return initClient(preparedQueryName, grpc.WithBalancer(lb))
}

func initClient(target string, grpcOptions ...grpc.DialOption) error {
	if echoClient != nil {
		return nil
	}

	dialoptions := make([]grpc.DialOption, 0)
	dialoptions = append(dialoptions, grpcOptions...)
	dialoptions = append(dialoptions, grpc.WithInsecure())
	dialoptions = append(dialoptions, grpc.WithBlock())

	// create a new client and wait to establish a connection before returnig
	var err error
	grpcClient, err = grpc.Dial(
		target,
		dialoptions...,
	)

	if err != nil {
		return fmt.Errorf("Setup error creating grpc client %s", err.Error())
	}

	echoClient = echo.NewEchoServiceClient(grpcClient)

	return nil
}
