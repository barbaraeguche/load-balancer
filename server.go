/*
* this is for individual backend server representation and management.
 */
package main

import (
	"net/url"
	"sync"
)

type Server struct {
	URL             *url.URL
	alive           bool
	weight          int // for weighted load balancing
	activeConnCount int
	maxAllowedConns int
	mutex           sync.RWMutex // for thread safety
}

// thread-safe check if server is up
func (s *Server) isAlive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.alive
}

// thread-safe setter for alive status
func (s *Server) setAlive(alive bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.alive = alive
}

// check if under connection limit
func (s *Server) canAcceptConnection() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.activeConnCount < s.maxAllowedConns
}

// track active connections
func (s *Server) incrementConnections() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.activeConnCount < s.maxAllowedConns {
		s.activeConnCount++
		return true
	}
	return false
}

// decrease connection count
func (s *Server) decrementConnections() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.activeConnCount > 0 {
		s.activeConnCount--
	}
}

// return server weight
func (s *Server) getWeight() int {
	return s.weight
}

// record response time for metrics
//func (s *Server) updateResponseTime() { //todo
//
//}
