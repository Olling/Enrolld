//package auth
//
//import (
//	"testing"
//	"github.com/Olling/Enrolld/utils/objects"
//)
//
//func TestCheckAccess(t *testing.T) {
////func CheckAccess(w http.ResponseWriter, r *http.Request, action string, server objects.Server) bool {
//	users := make(map[string]objects.User)
//
//	auth1 := objects.Authorization{UrlRegex: ".*", ServerIDRegex: ".*", GroupRegex: ".*", Actions: []string{"PUT", "GET", "POST", "DELETE"}}
//	user1 := objects.User{Password: "1234", Encrypted: false, Authorizations: []objects.Authorization{auth1}}
//	users["user1"] = user1
//
//	auth2 := objects.Authorization{UrlRegex: "/server", ServerIDRegex: ".*", GroupRegex: ".*", Actions: []string{"GET"}}
//	auth22 := objects.Authorization{UrlRegex: "/overwrite", ServerIDRegex: ".*", GroupRegex: ".*", Actions: []string{"POST"}}
//	user2 := objects.User{Password: "1234", Encrypted: false, Authorizations: []objects.Authorization{auth2, auth22}}
//	users["user2"] = user2
//
//	auth3 := objects.Authorization{UrlRegex: "/server", ServerIDRegex: "server1", GroupRegex: ".*", Actions: []string{"GET"}}
//	user3 := objects.User{Password: "1234", Encrypted: false, Authorizations: []objects.Authorization{auth3}}
//	users["user3"] = user3
//
//	var tests = []struct {
//		input []objects.Server
//		result bool
//	}{
//		{
//			input: []objects.Server{server1,server2},
//		},
//
//
//
//
//
//}
//
//

//func authenticate(w http.ResponseWriter, r *http.Request, username string, password string) bool {
//func authorize(url string, username string, server objects.Server, action string) bool {
