package dataaccess

import (
	"testing"
	"github.com/Olling/Enrolld/utils/objects"
)

func TestActiveServers(t *testing.T) {
	if ActiveServers != nil {
		t.Error("Test failed: ActiveServers was not nil")
	}

	var server1 objects.Server
	server1.ServerID = "server1"
	server1.IP = "127.0.0.1"
	server1.LastSeen = "2020-02-20 20:00:00.423709525 +0100 CET m=+1935960.050243285"
	server1.NewServer = true
	server1.Groups = []string{"group1", "group2"}

	if IsServerActive(server1.ServerID) {
		t.Error("Test failed: Got an inactive server as active")
	}

	MarkServerActive(server1)

	if len(ActiveServers) != 1 {
		t.Error("Test failed: MarkServerActive failed to activate server")
	}

	if ! IsServerActive(server1.ServerID) {
		t.Error("Test failed: Failed to find active server")
	}

	MarkServerInactive(server1)

	if len(ActiveServers) != 0 {
		t.Error("Test failed: MarkServerActive failed deactivate server")
	}

	if IsServerActive(server1.ServerID) {
		t.Error("Test failed: Got an inactive server as active")
	}
}
