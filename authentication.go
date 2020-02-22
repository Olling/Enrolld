package main

import (
	"os"
	"fmt"
	"net/url"
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"net/http/httputil"
	"github.com/olling/slog"
	"golang.org/x/crypto/bcrypt"
)


//var (
//	configuration *Configuration
//	configurationFile string
//)
//
//
//type Configuration struct {
//	Address string
//	LogLevel int
//	ProxyTargets []ProxyTarget
//	TlsEnabled bool
//	TlsAddress string
//	TlsKey string
//	TlsCert string
//	ClientCert_Enabled bool
//	ClientCert_CA string
//}

type BasicAuthCredential struct {
	Encrypted bool
	Username string
	Password string
}


func createTlsConfig(ClientCert_CA string) (tlsConfig tls.Config) {
	if !configuration.ClientCert_Enabled {
		return tlsConfig
	}

	certBytes, err := ioutil.ReadFile(ClientCert_CA)
	if err != nil {
		slog.PrintFatal("Unable to read:", ClientCert_CA, err)
	}

	certPool := x509.NewCertPool()
	if ok := certPool.AppendCertsFromPEM(certBytes); !ok {
		slog.PrintError("Unable to add cert to certpool:", ClientCert_CA)
	}

	tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	tlsConfig.ClientCAs = certPool
	tlsConfig.PreferServerCipherSuites = true
	tlsConfig.BuildNameToCertificate()

	return tlsConfig
}


func Authenticate (Username string, Password string) bool {
	authenticated := false
	for _, cred := range Credentials {
		if cred.Username != username {continue}

		if cred.Encrypted {
			if err := bcrypt.CompareHashAndPassword([]byte(cred.Password), []byte(password)); err == nil {
				authenticated = true
			}
		} else {
			if cred.Password == password {
				authenticated = true
			}
		}
	}

	return authenticated
}


func BasicAuthenticate(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	username, password, authOK := r.BasicAuth()
	if authOK == false {
		return false
	}

	return Authenticate(username, password)
}

//	if ! basicAuthenticate(w, r, proxyTarget) {
//		http.Error(w, "Not authorized", 401)
//		return
//	}
