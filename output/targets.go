package output

import (
	"io"
	"fmt"
	"sort"
	"strings"
	"crypto/sha1"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/dataaccess/config"
)

type TargetList struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func serverToTargetList(server utils.Server) (label string, entry TargetList) {
	if config.Configuration.TargetsPort != "" {
		server.ServerID = server.ServerID + ":" + config.Configuration.TargetsPort
	}

	entry.Targets = []string{server.ServerID}

	if server.Properties != nil {
		entry.Labels = server.Properties
	} else {
		entry.Labels = make(map[string]string)
	}

	entry.Labels["groups"] = strings.Join(server.Groups, ", ")

	if len(entry.Labels) == 0 {
		label = "nolabels"
	} else {
		sha1calc := sha1.New()

		var keys []string
		for key, _ := range entry.Labels {
			keys = append(keys, key)
		}
		sort.Strings(keys)

		for _, key := range keys {
			io.WriteString(sha1calc, key+":"+entry.Labels[key])
		}
		label = fmt.Sprintf("%x", sha1calc.Sum(nil))
	}

	return label, entry
}


func GetTargetsInJSON(servers []utils.Server) (string, error) {
	entriesmap := make(map[string]TargetList)

	for _, server := range servers {
		label, entry := serverToTargetList(server)

		_, keyexists := entriesmap[label]
		if keyexists {
			tempentry := entriesmap[label]
			tempentry.Targets = append(tempentry.Targets, entry.Targets...)
			entriesmap[label] = tempentry
		} else {
			entriesmap[label] = entry
		}
	}

	var entries []TargetList
	for _, value := range entriesmap {
		entries = append(entries, value)
	}
	entriesjson, err := utils.StructToJson(entries)

	return entriesjson, err
}
