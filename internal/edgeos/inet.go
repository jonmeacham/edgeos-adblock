package edgeos

import (
	"net"
	"strconv"
	"time"
)

// ChkWeb returns true if DNS is working
func ChkWeb(site string, port int) bool {
	target := net.JoinHostPort(site, strconv.Itoa(port))
	timeOut := 3 * time.Second
	conn, err := net.DialTimeout("tcp", target, timeOut)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}
