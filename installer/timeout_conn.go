package main

import (
	"net"
	"time"
)

// create custom net.Conn with read timeout
type timeoutConn struct {
	net.Conn
	timeout time.Duration
}

func (c *timeoutConn) Read(b []byte) (int, error) {
	c.Conn.SetReadDeadline(time.Now().Add(c.timeout))
	return c.Conn.Read(b)
}

func NewTimeoutConn(conn net.Conn, timeout time.Duration) net.Conn {
	return &timeoutConn{
		Conn:    conn,
		timeout: timeout,
	}
}
