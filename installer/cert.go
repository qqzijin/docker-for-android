package main

import (
	"crypto/x509"

	"github.com/jannson/gocertifi"
)

var rootCAs *x509.CertPool

func RootCAsGlobal() *x509.CertPool {
	return rootCAs
}

func init() {
	rootCAs, _ = gocertifi.CACerts()
}
