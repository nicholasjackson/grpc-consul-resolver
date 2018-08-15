package catalog

import (
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

type MockConsulAgent struct {
	mock.Mock
}

func (a *MockConsulAgent) ConnectCARoots(q *api.QueryOptions) (*api.CARootList, *api.QueryMeta, error) {
	args := a.Called(q)

	return args.Get(0).(*api.CARootList), nil, args.Error(2)
}
