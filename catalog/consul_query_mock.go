package catalog

import (
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

// MockConsulPreparedQuery is a mock impementation of the ConsulHealth interface used for testing
type MockConsulPreparedQuery struct {
	mock.Mock
}

// Service is used to query health information along with service info
// for a given service. It can optionally do server-side filtering on a tag
// or nodes with passing health checks only.
func (m *MockConsulPreparedQuery) Execute(queryIDOrName string, q *api.QueryOptions) (
	resp *api.PreparedQueryExecuteResponse, meta *api.QueryMeta, err error) {

	args := m.Called(queryIDOrName, q)

	resp = nil
	meta = nil
	err = args.Error(2)

	if e := args.Get(0); e != nil {
		resp = e.(func() *api.PreparedQueryExecuteResponse)()
	}

	if m := args.Get(1); m != nil {
		meta = m.(*api.QueryMeta)
	}

	return
}
