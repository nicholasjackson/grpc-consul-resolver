package catalog

import (
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/mock"
)

// MockConsulHealth is a mock impementation of the ConsulHealth interface used for testing
type MockConsulHealth struct {
	mock.Mock
}

// Service is used to query health information along with service info
// for a given service. It can optionally do server-side filtering on a tag
// or nodes with passing health checks only.
func (m *MockConsulHealth) Service(service, tag string, passingOnly bool, q *api.QueryOptions) (entries []*api.ServiceEntry, meta *api.QueryMeta, err error) {

	args := m.Called(service, tag, passingOnly, q)

	entries = nil
	meta = nil
	err = args.Error(2)

	if e := args.Get(0); e != nil {
		entries = e.(func() []*api.ServiceEntry)()
	}

	if m := args.Get(1); m != nil {
		meta = m.(*api.QueryMeta)
	}

	return
}

// Connect is TODO
func (m *MockConsulHealth) Connect(service, tag string, passingOnly bool, q *api.QueryOptions) (
	entries []*api.ServiceEntry, meta *api.QueryMeta, err error) {

	args := m.Called(service, tag, passingOnly, q)

	entries = nil
	meta = nil
	err = args.Error(2)

	if e := args.Get(0); e != nil {
		entries = e.(func() []*api.ServiceEntry)()
	}

	if m := args.Get(1); m != nil {
		meta = m.(*api.QueryMeta)
	}

	return
}
