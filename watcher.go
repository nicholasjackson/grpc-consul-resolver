package resolver

import (
	"fmt"
	"sync/atomic"
	"time"

	"github.com/hashicorp/consul/api"
	"google.golang.org/grpc/naming"
)

// ConsulWatcher is a service catalog watcher
type ConsulWatcher struct {
	client       ConsulHealth
	update       time.Duration
	service      string
	addressCache []string
	running      uint32
}

// NewConsulWatcher creates and returns a ConsulWatcher with the given parameters
func NewConsulWatcher(service string, c ConsulHealth, watchInterval time.Duration) *ConsulWatcher {
	return &ConsulWatcher{c, watchInterval, service, make([]string, 0), 1}
}

// Next blocks until an update or error happens. It may return one or more
// updates. The first call should get the full set of the results. It should
// return an error if and only if Watcher cannot recover.
func (c *ConsulWatcher) Next() ([]*naming.Update, error) {
	for atomic.LoadUint32(&c.running) == 1 {
		se, _, err := c.client.Service(c.service, "", true, nil)
		if err != nil {
			return nil, err
		}

		up, err := c.buildUpdate(se)

		if len(up) > 0 {
			return up, err
		}

		time.Sleep(c.update)
	}

	return nil, nil
}

// Close closes the Watcher.
func (c *ConsulWatcher) Close() {
	atomic.StoreUint32(&c.running, 0)
}

func (c *ConsulWatcher) buildUpdate(ses []*api.ServiceEntry) ([]*naming.Update, error) {
	nu := make([]*naming.Update, 0)
	nc := c.addressCache

	// check additions
	for _, se := range ses {
		addr := fmt.Sprintf("%s:%d", se.Service.Address, se.Service.Port)
		// does this address already exist in the cache?
		if !cacheContains(addr, c.addressCache) {
			nc = append(nc, addr)

			n := &naming.Update{
				Op:   naming.Add,
				Addr: addr,
			}

			nu = append(nu, n)
		}
	}

	// check deletions
	for _, a := range nc {
		if !serviceEntryContains(a, ses) {
			n := &naming.Update{
				Op:   naming.Delete,
				Addr: a,
			}

			nu = append(nu, n)
		}
	}

	c.addressCache = nc
	return nu, nil
}

func cacheContains(s string, in []string) bool {
	for _, i := range in {
		if s == i {
			return true
		}
	}

	return false
}

func serviceEntryContains(s string, in []*api.ServiceEntry) bool {
	for _, i := range in {
		if fmt.Sprintf("%s:%d", i.Service.Address, i.Service.Port) == s {
			return true
		}
	}

	return false
}
