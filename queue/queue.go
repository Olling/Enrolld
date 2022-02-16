package queue

import (
	"os"
	"fmt"
	"net"
	"errors"
	"regexp"
	"os/exec"
	"strings"
	"encoding/json"
	"github.com/Olling/slog"
	"github.com/Olling/Enrolld/utils/objects"
	"github.com/Olling/Enrolld/dataaccess/config"
)

type queue struct {
}

var (
	Configuration configuration
)


func StructToJson(s interface{}) (string, error) {
	bytes, marshalErr := json.MarshalIndent(s, "", "\t")
	return string(bytes), marshalErr
}
