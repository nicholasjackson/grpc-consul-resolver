package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/hashicorp/consul/api"
	echo "github.com/nicholasjackson/grpc-consul-resolver/functional_tests/grpc"
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
		port := rand.Intn(1000) + 6000

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
