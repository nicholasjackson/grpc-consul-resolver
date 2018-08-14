package catalog

import (
	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

type ServiceEntry struct {
	Addr    string
	CertURI *connect.SpiffeIDService
}

type Query interface {
	Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error)
}

type MockQuery struct {
	mock.Mock
}

func (m *MockQuery) Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error) {
	args := m.Called(name, options)

	if s := args.Get(0); s != nil {
		entries := s.(func() []ServiceEntry)()
		return entries, nil
	}

	return nil, args.Error(1)
}
