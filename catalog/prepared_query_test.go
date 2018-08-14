package catalog

import (
	"testing"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var queryMock *MockConsulPreparedQuery
var srs *api.PreparedQueryExecuteResponse

func testGetResponse() *api.PreparedQueryExecuteResponse {
	return srs
}

func setupPreparedQueryTests(t *testing.T) *PreparedQuery {
	srs = &api.PreparedQueryExecuteResponse{}
	srs.Nodes = []api.ServiceEntry{
		api.ServiceEntry{
			Service: &api.AgentService{
				Address: "localhost",
				Port:    8080,
			},
		},
	}

	queryMock = &MockConsulPreparedQuery{}
	queryMock.On("Execute", mock.Anything, mock.Anything).Return(testGetResponse, nil, nil)

	return NewPreparedQuery(queryMock)
}

func TestExecutePreparedQueryReturnsEntriesWhenServiceAddress(t *testing.T) {
	sq := setupPreparedQueryTests(t)

	entries, err := sq.Execute("localhost", nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "localhost:8080", entries[0].Addr)
}

func TestExecutePreparedQueryReturnsEntriesWhenNodeAddress(t *testing.T) {
	sq := setupPreparedQueryTests(t)
	srs.Nodes[0].Service.Address = ""
	srs.Nodes[0].Node = &api.Node{
		Address: "node",
	}

	entries, err := sq.Execute("node", nil)

	assert.NoError(t, err)
	assert.Len(t, entries, 1)
	assert.Equal(t, "node:8080", entries[0].Addr)
}
