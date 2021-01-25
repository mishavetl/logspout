package tcp

import (
	"net"
	"time"
	"syscall"
	"log"
	
	"github.com/gliderlabs/logspout/adapters/raw"
	"github.com/gliderlabs/logspout/router"
)

func init() {
	router.AdapterTransports.Register(new(tcpTransport), "tcp")
	// convenience adapters around raw adapter
	router.AdapterFactories.Register(rawTCPAdapter, "tcp")
}

func rawTCPAdapter(route *router.Route) (router.LogAdapter, error) {
	route.Adapter = "raw+tcp"
	return raw.NewRawAdapter(route)
}

type tcpTransport int

func SetupKeepAlive(conn *net.TCPConn, idleTime time.Duration, retryInterval int, pingAmount int) {
	log.Println("Setting keepalive")

	conn.SetKeepAlive(true)
	conn.SetKeepAlivePeriod(idleTime)

	// Getting the file handle of the socket
	sockFile, sockErr := conn.File()
	if sockErr == nil {
		// got socket file handle. Getting descriptor.
		fd := int(sockFile.Fd())
		// Ping amount
		err := syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPCNT, pingAmount)
		if err != nil {
			log.Println("on setting keepalive probe count", err.Error())
		}
		// Retry interval
		err = syscall.SetsockoptInt(fd, syscall.IPPROTO_TCP, syscall.TCP_KEEPINTVL, retryInterval)
		if err != nil {
			log.Println("on setting keepalive retry interval", err.Error())
		}
		// don't forget to close the file. No worries, it will *not* cause the connection to close.
		sockFile.Close()
	} else {
		log.Println("on setting socket keepalive", sockErr.Error())
	}
}

func (t *tcpTransport) Dial(addr string, options map[string]string) (net.Conn, error) {
	raddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTCP("tcp", nil, raddr)
	if err != nil {
		return nil, err
	}
	idleTime, _ := time.ParseDuration("30s")
	SetupKeepAlive(conn, idleTime, 5, 3)
	return conn, nil
}
