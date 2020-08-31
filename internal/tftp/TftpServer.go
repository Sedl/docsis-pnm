package tftp

import (
	"errors"
	"github.com/pin/tftp"
	"io"
	"log"
	"net"
	"sync"
	"time"
)

const Mebibyte = 1048576

var TimeoutError = errors.New("timeout while waiting for file")

type Server struct {
	Server     *tftp.Server
	requests   map[string]chan *SafeBuffer
	mutex      sync.RWMutex
	ExternalIp net.IP
}

// writeHandler handles TFTP write requests. It sends a "permission denied" to the client if we don't expect a file we don't
// wait for.
func (s *Server) writeHandler(filename string, wt io.WriterTo) error {
	s.mutex.RLock()
	channel, ok := s.requests[filename]
	s.mutex.RUnlock()

	if !ok {
		return errors.New("permission denied")
	}
	buf := NewSafeBuffer(4 * Mebibyte)
	byteCount, err := wt.WriteTo(buf)
	log.Printf("TFTP: received file %s (%d bytes)\n", filename, byteCount)
	channel <- buf
	return err
}

// ListenAndServe is a wrapper around tftp.ListenAndServe
func (s *Server) ListenAndServe(addr string) error {
	log.Printf("starting TFTP server using external IP address %s\n", s.ExternalIp)
	return s.Server.ListenAndServe(addr)
}

// TODO implement Context
// WaitForFile blocks until we receive a file with the given name via TFTP, or the time runs out
func (s *Server) WaitForFile(filename string, timeout time.Duration) (*SafeBuffer, error) {
	channel := make(chan *SafeBuffer)
	s.mutex.Lock()
	s.requests[filename] = channel
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		delete(s.requests, filename)
		s.mutex.Unlock()
	}()

	timer := time.NewTimer(timeout)
	select {
	case buffer := <-channel:
		return buffer, nil
	case <-timer.C:
		return nil, TimeoutError
	}
}

func NewServer(externalIp net.IP) *Server {
	server := &Server{
		requests: make(map[string]chan *SafeBuffer),
	}

	// check if it is a "real" IPv4 address or (as by default in golang after parsing) an IPv4-mapped IPv6 address
	ipv4 := externalIp.To4()
	if ipv4 != nil {
		server.ExternalIp = ipv4
	} else {
		server.ExternalIp = externalIp
	}

	server.Server = tftp.NewServer(nil, server.writeHandler)
	server.Server.SetTimeout(time.Second * 5)
	return server
}
