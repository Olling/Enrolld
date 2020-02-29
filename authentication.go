package main

import (
	"net/http"
	"io/ioutil"
	"crypto/tls"
	"crypto/x509"
	"github.com/Olling/slog"
	"golang.org/x/crypto/bcrypt"
)


type Credential struct {
	Encrypted bool
	Username string
	Password string
}


func CreateTlsConfig(ClientCert_CA string) (tlsConfig tls.Config) {
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


func Authenticate (username string, password string, credentials []Credential) bool {
	authenticated := false
	for _, cred := range credentials {
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


func BasicAuthenticate(w http.ResponseWriter, r *http.Request, credentials []Credential) bool {
	w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)

	username, password, authOK := r.BasicAuth()
	if authOK == false {
		return false
	}

	return Authenticate(username, password, credentials)
}

//	if ! basicAuthenticate(w, r, proxyTarget) {
//		http.Error(w, "Not authorized", 401)
//		return
//	}
