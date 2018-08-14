package catalog

import "github.com/hashicorp/consul/api"

// ConsulHealth defines an interface which adheres to the required functions from
// the github.com/hashicorp/consul/api Health struct
type ConsulHealth interface {
	Service(service, tag string, passingOnly bool, q *api.QueryOptions) (
		[]*api.ServiceEntry, *api.QueryMeta, error)

	Connect(service, tag string, passingOnly bool, q *api.QueryOptions) (
		[]*api.ServiceEntry, *api.QueryMeta, error)
}

// ConsulPreparedQuery defines an interface which adhers to the required functions from
// the github.com/hashicorp/consul/api PreparedQuery struct
type ConsulPreparedQuery interface {
	Execute(queryIDOrName string, q *api.QueryOptions) (
		*api.PreparedQueryExecuteResponse, *api.QueryMeta, error)
}
