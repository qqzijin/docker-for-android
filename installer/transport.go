package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"time"
)

const (
	ISE_HTTP_LOG = "ISE_HTTP_LOG"
)

func CreateTimeoutTransport(timeout time.Duration) *http.Transport {
	certPool := RootCAsGlobal()
	cfg := &tls.Config{RootCAs: certPool}
	return &http.Transport{
		TLSClientConfig: cfg,
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, _, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			tconn, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				return nil, err
			}
			return NewTimeoutConn(tls.Client(tconn, &tls.Config{
				ServerName: host,
				RootCAs:    certPool,
			}), timeout), nil
		},
	}
}

type LogTransport struct {
	lower http.RoundTripper
}

func (lt *LogTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqd, err0 := httputil.DumpRequestOut(req, false)
	if err0 != nil {
		return nil, err0
	}
	timeStart := time.Now().Unix()
	resp, err := lt.lower.RoundTrip(req)
	if err != nil {
		fmt.Println(string(reqd))
		fmt.Println("Failed ", err)
		fmt.Println()
		return resp, err
	}
	respd, err0 := httputil.DumpResponse(resp, false)
	if err0 != nil {
		return nil, err0
	}
	fmt.Println(string(reqd))
	fmt.Println(string(respd))
	fmt.Println("Success in ", time.Now().Unix()-timeStart, " seconds")
	return resp, err
}

func CreateLogTransport(lower http.RoundTripper) http.RoundTripper {
	if os.Getenv(ISE_HTTP_LOG) == "1" {
		return &LogTransport{
			lower: lower,
		}
	} else {
		return lower
	}
}
