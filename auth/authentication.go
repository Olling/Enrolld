package auth

import (
	"regexp"
	"net/http"
	"github.com/Olling/slog"
	"golang.org/x/crypto/bcrypt"
	"github.com/Olling/Enrolld/utils"
	"github.com/Olling/Enrolld/config"
	"github.com/Olling/Enrolld/dataaccess/fileio"
)


var (
	Users		map[string]User
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


func Initialize() {
	slog.PrintDebug("Initializing Authentication")
	err := fileio.LoadFromFile(&Users, config.Configuration.FileBackendDirectory + "/auth.json")
	if err != nil {
		slog.PrintError("Failed to load auth.json", err)
	}
}


func CheckAccess(w http.ResponseWriter, r *http.Request, action string, server utils.Server) bool {
	// Anonymous Access
	if _,anonAccess := Users["anonymous"]; anonAccess {
		for _,authorization := range Users["anonymous"].Authorizations {
			if matched,_ := regexp.MatchString(authorization.UrlRegex, r.URL.String()); matched {
				if utils.StringExistsInArray(authorization.Actions, action) {
					return true
				}
			}
		}
	}

	// Authenticate User
	username, password, authOK := r.BasicAuth()
	if authOK == false {
		return false
	}

	authenticated := authenticate(w, r, username, password)
	if !authenticated {
		return authenticated
	}

	// Authorize User
	return authorize(r.URL.String(), username, server, action)
}


func authenticate(w http.ResponseWriter, r *http.Request, username string, password string) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	authenticated := false
	for name,user := range Users {
		if name != username {continue}

		if user.Encrypted {
			if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err == nil {
				authenticated = true
			}
		} else {
			if user.Password == password {
				authenticated = true
			}
		}
	}

	return authenticated
}


func authorize(url string, username string, server utils.Server, action string) bool {
	if _,userAccess := Users[username]; !userAccess {
		return false
	}

	match_url := false

	var auth Authorization
	for _,authorization := range Users[username].Authorizations {
		if matched,_ := regexp.MatchString(authorization.UrlRegex, url); matched {
			if utils.StringExistsInArray(authorization.Actions, action) {
				auth = authorization
				match_url = true
				continue
			}
		}
	}

	if !match_url {
		return false
	}

	if server.ServerID == "" {
		return true
	}

	if matched,_ := regexp.MatchString(auth.ServerIDRegex, server.ServerID); !matched {
		return false
	}

	match := false
	for _, group := range server.Groups {
		if matched,_ := regexp.MatchString(auth.GroupRegex, group); matched {
			match = true
		}
	}

	return match
}
