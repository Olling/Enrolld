package objects

import (
	"time"
)

type User struct {
	Password		string
	Encrypted		bool
	Authorizations		[]Authorization
}

type Authorization struct {
	Actions		[]string
	UrlRegex	string
	ServerIDRegex	string
	GroupRegex	string
}

type KeyValue struct {
	Key			string
	Value			string
}

type Overwrite struct {
	ServerIDRegexp		string
	GroupRegexp		string
	PropertiesRegexp	KeyValue
	Groups			[]string
	Properties		map[string]string
}

type Script struct {
	Description	string
	Path		string
	Timeout		int
}

type Server struct {
	ServerID	string
	IP		string
	LastSeen	string
	NewServer	bool `json:"NewServer,omitempty"`
	Groups		[]string
	Properties	map[string]string
}

type Job struct {
	Server Server
	WorkerID string
	JobID string
	Timestamp time.Time
}
