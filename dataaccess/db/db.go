package db

import (
	"github.com/Olling/Enrolld/utils/objects"
)

func LoadAuthentication(users *map[string]objects.User) error {
	return nil
}

func DeleteServer(serverName string) error {
	return nil
}

func LoadOverwrites(overwrites *interface{}) {
	//Write to overwrites
}

func SaveOverwrites(overwrites interface{}) {
}

func AddOverwrites(server *objects.Server, overwrites map[string]objects.Overwrite) {
}

func GetServers(overwrites map[string]objects.Overwrite) ([]objects.Server, error) {
	return []objects.Server{}, nil
}

func ServerExist(serverID string) bool {
	return false
}

func RemoveServer(serverID string) error {
	//if err == nil {
	//	metrics.ServersDeleted.Inc()
	//}
	return nil
}

func GetServer(serverID string, overwrites map[string]objects.Overwrite) (server objects.Server, err error) {
	return server, err
}

func UpdateServer(server objects.Server, isNewServer bool) error {
	return nil
}

func IsServerActive(serverID string) bool {
	return false
}

func MarkServerActive(server objects.Server) error {
	return nil
}

func MarkServerInactive(server objects.Server) error {
	return nil
}

func GetServerCount() float64 {
	return 0
}

func GetFilteredServersList(groups []string, properties map[string]string) ([]objects.Server, error) {
	return []objects.Server{}, nil
}
