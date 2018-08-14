package catalog

import (
	"fmt"

	"github.com/hashicorp/consul/api"
)

type ServiceQuery struct {
	client ConsulHealth
}

func NewServiceQuery(client ConsulHealth) *ServiceQuery {
	return &ServiceQuery{client}
}

func (s *ServiceQuery) Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error) {
	ses := make([]ServiceEntry, 0)

	services, _, err := s.client.Service(name, "", true, options)
	if err != nil {
		return nil, err
	}

	for _, s := range services {
		se := ServiceEntry{}
		se.Addr = buildAddress(s)

		ses = append(ses, se)
	}

	return ses, nil
}

func buildAddress(se *api.ServiceEntry) string {
	if se.Service.Address != "" {
		return fmt.Sprintf("%s:%d", se.Service.Address, se.Service.Port)
	}

	return fmt.Sprintf("%s:%d", se.Node.Address, se.Service.Port)
}
