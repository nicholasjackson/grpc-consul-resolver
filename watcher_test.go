package resolver

import (
	"fmt"
	"testing"
	"time"

	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/naming"
)

var healthMock *MockConsulHealth
var ses []*api.ServiceEntry

func getServices() []*api.ServiceEntry {
	return ses
}

func setupWatcher(t *testing.T) *ConsulWatcher {
	ses = make([]*api.ServiceEntry, 1)
	ses[0] = &api.ServiceEntry{
		Service: &api.AgentService{
			Address: "localhost",
			Port:    8080,
		},
	}

	healthMock = &MockConsulHealth{}
	healthMock.On("Service", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(getServices, nil, nil)

	return NewConsulWatcher("test", healthMock, 1*time.Millisecond)
}

func TestNewConsulWatcherReturnsWatcher(t *testing.T) {
	w := NewConsulWatcher("test", &MockConsulHealth{}, 1*time.Second)

	assert.NotNil(t, w)
}

func TestNextReturnsErrorWhenConsulError(t *testing.T) {
	w := setupWatcher(t)
	healthMock.ExpectedCalls = make([]*mock.Call, 0)
	healthMock.On("Service", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, nil, fmt.Errorf("Boom"))

	_, err := w.Next()

	assert.NotNil(t, err, "Should have returned an error")
}

func TestNextReturnsInitialUpdatesFromConsul(t *testing.T) {
	w := setupWatcher(t)

	nu, err := w.Next()

	assert.NoError(t, err)
	assert.Len(t, nu, 1, "Should have returned 1 update")
	assert.Equal(t, "localhost:8080", nu[0].Addr)
	assert.Equal(t, naming.Add, nu[0].Op)
}

func TestNextReturnsUpdatesContainingAddedItemsFromConsul(t *testing.T) {
	w := setupWatcher(t)
	w.Next()
	ses = append(ses, &api.ServiceEntry{
		Service: &api.AgentService{
			Address: "localhost",
			Port:    8090,
		},
	})

	nu, err := w.Next()

	assert.NoError(t, err)
	healthMock.AssertNumberOfCalls(t, "Service", 2)
	assert.Len(t, nu, 1, "Should have returned 1 updates")

	assert.Equal(t, "localhost:8090", nu[0].Addr)
	assert.Equal(t, naming.Add, nu[0].Op)
}

func TestNextReturnsUpdatesContainingDeletedItemsFromConsul(t *testing.T) {
	w := setupWatcher(t)
	w.Next()
	ses = make([]*api.ServiceEntry, 0)

	nu, err := w.Next()

	assert.NoError(t, err)
	healthMock.AssertNumberOfCalls(t, "Service", 2)
	assert.Len(t, nu, 1, "Should have returned 1 updates")

	assert.Equal(t, "localhost:8080", nu[0].Addr)
	assert.Equal(t, naming.Delete, nu[0].Op)
}

func TestNextBlocksWhenNoChangesFromConsul(t *testing.T) {
	w := setupWatcher(t)
	w.Next()

	timeOut := make(chan bool)

	// test after n seconds
	time.AfterFunc(2*time.Millisecond, func() {
		timeOut <- true
	})

	go func() {
		w.Next()
		t.Fatal("Next should block and never return")
	}()

	// check that the next call blocks for n itterations when no changes from consul
	<-timeOut
	w.Close() // stop the watcher
	healthMock.AssertNumberOfCalls(t, "Service", 3)
}
