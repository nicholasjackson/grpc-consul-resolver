package catalog

import (
	"fmt"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

// ServiceEntry describes the details for service resolution, CertURI will be
// null unless the Service is a Consul Connect service.
type ServiceEntry struct {
	Addr    string
	CertURI connect.CertURI
}

// Query defines an interface for service discovery methods to implement,
// like Prepared Query, Consul Service catalog
type Query interface {
	Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error)
}

// MockQuery is a mock implementation of service resolution for use in tests
type MockQuery struct {
	mock.Mock
}

// Execute and return mock data for tests
func (m *MockQuery) Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error) {
	args := m.Called(name, options)

	if s := args.Get(0); s != nil {
		entries := s.(func() []ServiceEntry)()
		return entries, nil
	}

	return nil, args.Error(1)
}

// helper function to build the address for the upstream service
func buildAddress(se *api.ServiceEntry) string {
	if se.Service.Address != "" {
		return fmt.Sprintf("%s:%d", se.Service.Address, se.Service.Port)
	}

	return fmt.Sprintf("%s:%d", se.Node.Address, se.Service.Port)
}
