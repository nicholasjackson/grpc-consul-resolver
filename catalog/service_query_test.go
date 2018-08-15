package catalog

import (
	"testing"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var healthMock *MockConsulHealth
var agentMock *MockConsulAgent
var ses []*api.ServiceEntry

func testGetServices() []*api.ServiceEntry {
	return ses
}

func setupServiceQueryTests(t *testing.T, useConnect bool) *ServiceQuery {
	ses = make([]*api.ServiceEntry, 1)
	ses[0] = &api.ServiceEntry{
		Service: &api.AgentService{
			Address:          "localhost",
			Port:             8080,
			ProxyDestination: "localhost:9999",
			Connect: &api.AgentServiceConnect{
				Native: false,
			},
			Service: "localhost:8081",
		},
		Node: &api.Node{
			Datacenter: "dc1",
		},
	}

	agentMock = &MockConsulAgent{}
	agentMock.On("ConnectCARoots", mock.Anything).Return(&api.CARootList{TrustDomain: "abc.com"}, nil, nil)

	healthMock = &MockConsulHealth{}
	healthMock.On("Service", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testGetServices, nil, nil)
	healthMock.On("Connect", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testGetServices, nil, nil)

	return &ServiceQuery{healthMock, agentMock, useConnect, ""}
}

func TestExecuteServiceQueryReturnsEntriesWhenServiceAddress(t *testing.T) {
	sq := setupServiceQueryTests(t, false)

	entries, err := sq.Execute("localhost", nil)

	healthMock.AssertCalled(t, "Service", mock.Anything, mock.Anything, true, mock.Anything)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "localhost:8080", entries[0].Addr)
}

func TestExecuteServiceQueryReturnsEntriesWhenNodeAddress(t *testing.T) {
	sq := setupServiceQueryTests(t, false)
	ses[0].Service.Address = ""
	ses[0].Node = &api.Node{
		Address: "node",
	}

	entries, err := sq.Execute("node", nil)

	healthMock.AssertCalled(t, "Service", mock.Anything, mock.Anything, true, mock.Anything)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "node:8080", entries[0].Addr)
}

func TestExecuteConnectServiceQueryReturnsValidCertURINotNative(t *testing.T) {
	sq := setupServiceQueryTests(t, true)

	entries, err := sq.Execute("localhost.service.connect", nil)

	healthMock.AssertCalled(t, "Connect", mock.Anything, mock.Anything, true, mock.Anything)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "localhost:8080", entries[0].Addr)

	spiffeID := entries[0].CertURI.(*connect.SpiffeIDService)
	assert.Equal(t, "default", spiffeID.Namespace)
	assert.Equal(t, "dc1", spiffeID.Datacenter)
	assert.Equal(t, "localhost:9999", spiffeID.Service)
}

func TestExecuteConnectServiceQueryReturnsValidCertURINative(t *testing.T) {
	sq := setupServiceQueryTests(t, true)
	ses[0].Service.Connect.Native = true

	entries, err := sq.Execute("localhost.service.connect", nil)

	healthMock.AssertCalled(t, "Connect", mock.Anything, mock.Anything, true, mock.Anything)
	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "localhost:8080", entries[0].Addr)

	spiffeID := entries[0].CertURI.(*connect.SpiffeIDService)
	assert.Equal(t, "default", spiffeID.Namespace)
	assert.Equal(t, "dc1", spiffeID.Datacenter)
	assert.Equal(t, "localhost:8081", spiffeID.Service)
	assert.Equal(t, "abc.com", spiffeID.Host)
}
