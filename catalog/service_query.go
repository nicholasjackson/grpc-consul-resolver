package catalog

import (
	"fmt"

	"github.com/hashicorp/consul/agent/connect"
	"github.com/hashicorp/consul/api"
)

// ServiceQuery implements the logic to lookup a service in Consul's Service Catalog
type ServiceQuery struct {
	client      ConsulHealth
	agent       ConsulAgent
	useConnect  bool // should we query the
	trustDomain string
}

// NewServiceQuery creates a new ServiceQuery struct configured with a Consul API
// Client
// Setting the useConnect parameter to true will query the Consul Connect service
// catalog and return the address to the Connect proxy associated with the service
func NewServiceQuery(client *api.Client, useConnect bool) *ServiceQuery {
	return &ServiceQuery{client.Health(), client.Agent(), useConnect, ""}
}

// Execute the query against the API and build a list of ServiceEntry structs
// which can be used by the resolver
func (s *ServiceQuery) Execute(name string, options *api.QueryOptions) ([]ServiceEntry, error) {
	ses := make([]ServiceEntry, 0)

	var services []*api.ServiceEntry
	var err error

	// Are we looking up the service in the standard service catalog or the connect
	// service catalog
	if s.useConnect {
		services, _, err = s.client.Connect(name, "", true, options)
	} else {
		services, _, err = s.client.Service(name, "", true, options)
	}

	if err != nil {
		return nil, err
	}

	for _, svc := range services {
		se := ServiceEntry{}
		se.Addr = buildAddress(svc)

		if s.useConnect {
			certURI, err := s.buildCert(svc)
			if err != nil {
				return nil, err
			}

			se.CertURI = certURI
		}

		ses = append(ses, se)
	}

	return ses, nil
}

func (s *ServiceQuery) buildCert(se *api.ServiceEntry) (connect.CertURI, error) {
	service := se.Service.Proxy.DestinationServiceName
	if se.Service.Connect != nil && se.Service.Connect.Native {
		service = se.Service.Service
	}

	if service == "" {
		// Shouldn't happen but to protect against bugs in agent API returning bad
		// service response...
		return nil, fmt.Errorf("not a valid connect service")
	}

	// if we have not trust domain fetch it
	if s.trustDomain == "" {
		r, _, err := s.agent.ConnectCARoots(nil)
		if err != nil {
			return nil, err
		}

		s.trustDomain = r.TrustDomain
	}

	// Generate the expected CertURI
	certURI := &connect.SpiffeIDService{
		Host:       s.trustDomain,
		Namespace:  "default",
		Datacenter: se.Node.Datacenter,
		Service:    service,
	}

	return certURI, nil
}
