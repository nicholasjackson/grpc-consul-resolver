package resolver

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	"google.golang.org/grpc/naming"
)

// ConsulWatcher is a service catalog watcher
type ConsulWatcher struct {
	query        catalog.Query
	update       time.Duration
	service      string
	addressCache map[string]catalog.ServiceEntry
	running      uint32
}

// NewConsulWatcher creates and returns a ConsulWatcher with the given parameters
func NewConsulWatcher(service string, q catalog.Query, watchInterval time.Duration) *ConsulWatcher {
	return &ConsulWatcher{q, watchInterval, service, make(map[string]catalog.ServiceEntry), 1}
}

// Next blocks until an update or error happens. It may return one or more
// updates. The first call should get the full set of the results. It should
// return an error if and only if Watcher cannot recover.
func (c *ConsulWatcher) Next() ([]*naming.Update, error) {
	for atomic.LoadUint32(&c.running) == 1 {
		se, err := c.query.Execute(c.service, nil)
		if err != nil {
			return nil, err
		}

		up, err := c.buildUpdate(se)

		if len(up) > 0 {
			return up, nil
		}

		time.Sleep(c.update)
	}

	return nil, nil
}

// Close closes the Watcher.
func (c *ConsulWatcher) Close() {
	atomic.StoreUint32(&c.running, 0)
}

func (c *ConsulWatcher) buildUpdate(ses []catalog.ServiceEntry) ([]*naming.Update, error) {
	nu := make([]*naming.Update, 0)

	// check additions
	for _, se := range ses {
		addr := se.Addr
		// does this address already exist in the cache?
		if _, ok := c.addressCache[addr]; ok != true {
			c.addressCache[addr] = se

			n := &naming.Update{
				Op:   naming.Add,
				Addr: addr,
			}

			nu = append(nu, n)
		}
	}

	// check deletions
	for k := range c.addressCache {
		if !serviceEntryContains(k, ses) {
			n := &naming.Update{
				Op:   naming.Delete,
				Addr: k,
			}

			nu = append(nu, n)
		}
	}

	return nu, nil
}

func serviceEntryContains(s string, in []catalog.ServiceEntry) bool {
	for _, i := range in {
		if i.Addr == s {
			return true
		}
	}

	return false
}

func buildAddress(se *api.ServiceEntry) string {
	if se.Service.Address != "" {
		return fmt.Sprintf("%s:%d", se.Service.Address, se.Service.Port)
	}

	return fmt.Sprintf("%s:%d", se.Node.Address, se.Service.Port)
}
