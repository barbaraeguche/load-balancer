/*
* this is for individual backend server representation and management.
 */
package main

import (
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL                 *url.URL
	alive               bool
	weight              int // for weighted load balancing
	activeConnCount     int
	maxAllowedConns     int
	lastHealthCheck     time.Time
	responseTimeHistory [100]time.Duration // for metrics
	responseTimeIndex   int                // circular buffer index
	mutex               sync.RWMutex       // for thread safety
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
	s.lastHealthCheck = time.Now()
}

// check if under connection limit
func (s *Server) canAcceptConnection() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.alive && s.activeConnCount < s.maxAllowedConns
}

// track active connections
func (s *Server) incrementConnections() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.alive && s.activeConnCount < s.maxAllowedConns {
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

// record response time for metrics using circular buffer
func (s *Server) updateResponseTime(responseTime time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.responseTimeHistory[s.responseTimeIndex] = responseTime
	s.responseTimeIndex = (s.responseTimeIndex + 1) % len(s.responseTimeHistory)
}

// get average response time from recent history
func (s *Server) getAverageResponseTime() time.Duration {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var count int
	var totalTime time.Duration

	for _, rt := range s.responseTimeHistory {
		if rt > 0 {
			totalTime += rt
			count++
		}
	}

	// none worth calculating
	if count == 0 {
		return 0
	}

	return totalTime / time.Duration(count)
}

// constructor for new server
func NewServer(serverUrl string, weight, maxConns int) (*Server, error) {
	parsedUrl, err := url.Parse(serverUrl)

	if err != nil {
		return nil, err
	}

	return &Server{
		URL:             parsedUrl,
		alive:           true,
		weight:          weight,
		maxAllowedConns: maxConns,
		lastHealthCheck: time.Now(),
	}, nil
}
