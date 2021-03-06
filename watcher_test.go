package resolver

import (
	"fmt"
	"testing"
	"time"

	"github.com/nicholasjackson/grpc-consul-resolver/catalog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/naming"
)

var queryMock *catalog.MockQuery
var ses []catalog.ServiceEntry

func getServices() []catalog.ServiceEntry {
	return ses
}

func setupWatcher(t *testing.T) *ConsulWatcher {
	ses = make([]catalog.ServiceEntry, 1)
	ses[0] = catalog.ServiceEntry{
		Addr: "localhost:8080",
	}

	queryMock = &catalog.MockQuery{}
	queryMock.On("Execute", mock.Anything, mock.Anything).Return(getServices, nil)

	return NewConsulWatcher("test", queryMock, 10*time.Millisecond)
}

func TestNewConsulWatcherReturnsWatcher(t *testing.T) {
	w := NewConsulWatcher("test", &catalog.MockQuery{}, 1*time.Second)

	assert.NotNil(t, w)
}

func TestNextReturnsErrorWhenConsulError(t *testing.T) {
	w := setupWatcher(t)
	queryMock.ExpectedCalls = make([]*mock.Call, 0)
	queryMock.On("Execute", mock.Anything, mock.Anything).Return(nil, fmt.Errorf("Boom"))

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
	ses = append(ses, catalog.ServiceEntry{
		Addr: "localhost:8090",
	})

	nu, err := w.Next()

	assert.NoError(t, err)
	queryMock.AssertNumberOfCalls(t, "Execute", 2)
	assert.Len(t, nu, 1, "Should have returned 1 updates")

	assert.Equal(t, "localhost:8090", nu[0].Addr)
	assert.Equal(t, naming.Add, nu[0].Op)
}

func TestNextReturnsUpdatesContainingDeletedItemsFromConsul(t *testing.T) {
	w := setupWatcher(t)
	w.Next()
	ses = make([]catalog.ServiceEntry, 0)

	nu, err := w.Next()

	assert.NoError(t, err)
	queryMock.AssertNumberOfCalls(t, "Execute", 2)
	assert.Len(t, nu, 1, "Should have returned 1 updates")

	assert.Equal(t, "localhost:8080", nu[0].Addr)
	assert.Equal(t, naming.Delete, nu[0].Op)
}

func TestNextBlocksWhenNoChangesFromConsul(t *testing.T) {
	w := setupWatcher(t)
	w.Next()

	timeOut := make(chan bool)

	// test after 3 iterations
	time.AfterFunc(31*time.Millisecond, func() {
		timeOut <- true
	})

	testComplete := false
	go func() {
		w.Next()
		assert.True(t, testComplete, "Next should not have returned before close was called")
	}()

	// check that the next call blocks for n itterations when no changes from consul
	<-timeOut
	testComplete = true
	w.Close()                         // stop the watcher
	time.Sleep(10 * time.Millisecond) // wait for exit as loop might be sleeping

	queryMock.AssertNumberOfCalls(t, "Execute", 4)
}
