package dataaccess

import (
	"sync"
	"time"
	"errors"
	"github.com/Olling/Enrolld/utils/objects"
)

var (
	SyncActiveMutex		sync.Mutex
	ActiveServers		map[string]time.Time
)

func IsServerActive(serverID string) bool {
	if ActiveServers == nil {
		return false
	}

	SyncActiveMutex.Lock()
	defer SyncActiveMutex.Unlock()

	_, exist := ActiveServers[serverID]
	return exist
}

func MarkServerActive(server objects.Server) error {
	if ActiveServers == nil {
		ActiveServers = make(map[string]time.Time)
	}

	SyncActiveMutex.Lock()
	defer SyncActiveMutex.Unlock()

	if _, ok := ActiveServers[server.ServerID]; !ok {
		ActiveServers[server.ServerID] = time.Now()
		return nil
	}
	return errors.New("Server is already active")
}

func MarkServerInactive(server objects.Server) error {
	if ActiveServers != nil {
		SyncActiveMutex.Lock()
		defer SyncActiveMutex.Unlock()

		if _, ok := ActiveServers[server.ServerID]; ok {
			delete(ActiveServers, server.ServerID)
		}
	}

	return nil
}
