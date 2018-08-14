package catalog

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var healthMock *MockConsulHealth
var ses []*api.ServiceEntry

func testGetServices() []*api.ServiceEntry {
	return ses
}

func setupServiceQueryTests(t *testing.T) *ServiceQuery {
	ses = make([]*api.ServiceEntry, 1)
	ses[0] = &api.ServiceEntry{
		Service: &api.AgentService{
			Address: "localhost",
			Port:    8080,
		},
	}

	healthMock = &MockConsulHealth{}
	healthMock.On("Service", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(testGetServices, nil, nil)

	return NewServiceQuery(healthMock)
}

func TestExecuteServiceQueryReturnsEntriesWhenServiceAddress(t *testing.T) {
	sq := setupServiceQueryTests(t)

	entries, err := sq.Execute("localhost", nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "localhost:8080", entries[0].Addr)
}

func TestExecuteServiceQueryReturnsEntriesWhenNodeAddress(t *testing.T) {
	sq := setupServiceQueryTests(t)
	ses[0].Service.Address = ""
	ses[0].Node = &api.Node{
		Address: "node",
	}

	entries, err := sq.Execute("node", nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "node:8080", entries[0].Addr)
}
