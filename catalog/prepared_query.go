package catalog

import "github.com/hashicorp/consul/api"

type PreparedQuery struct {
	client ConsulPreparedQuery
}

func NewPreparedQuery(client ConsulPreparedQuery) *PreparedQuery {
	return &PreparedQuery{client}
}

func (s *PreparedQuery) Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error) {
	pqr, _, err := s.client.Execute(name, options)
	if err != nil {
		return nil, err
	}

	ses := make([]ServiceEntry, 0)
	for _, se := range pqr.Nodes {
		s := ServiceEntry{
			Addr: buildAddress(&se),
		}

		ses = append(ses, s)
	}

	return ses, nil
}
