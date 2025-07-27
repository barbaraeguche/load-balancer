/*
*
purpose: individual backend server representation and management.
*/
package main

import (
	"net/url"
	"sync"
	"time"
)

type Server struct {
	URL                  *url.URL
	Alive                bool
	Weight               int // for weighted load balancing
	ActiveConnsCount     int
	MaxAllowedConns      int
	LastHealthCheck      time.Time
	ResponseTimeHistory  [100]time.Duration // for metrics
	ResponseHistoryIndex int
	mutex                sync.RWMutex
}

// IsAlive thread-safe check if server is up
func (s *Server) IsAlive() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.Alive
}

// SetAlive thread-safe setter for alive status
func (s *Server) SetAlive(alive bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.Alive = alive
	s.LastHealthCheck = time.Now()
}

// CanAcceptConnections check if under connection limit
func (s *Server) CanAcceptConnections() bool {
	serverIsAlive := s.IsAlive()

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return serverIsAlive && s.ActiveConnsCount < s.MaxAllowedConns
}

// IncrementConnections increase connection count
func (s *Server) IncrementConnections() bool {
	serverCanAcceptConns := s.CanAcceptConnections()

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if serverCanAcceptConns {
		s.ActiveConnsCount++
		return true
	}
	return false
}

// DecrementConnections decrease connection count
func (s *Server) DecrementConnections() bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if s.ActiveConnsCount > 0 {
		s.ActiveConnsCount--
		return true
	}
	return false
}

// GetWeight return server weight
func (s *Server) GetWeight() int {
	return s.Weight
}

// UpdateResponseTime record response time for metrics
func (s *Server) UpdateResponseTime(responseTime time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	s.ResponseTimeHistory[s.ResponseHistoryIndex] = responseTime
	s.ResponseHistoryIndex = (s.ResponseHistoryIndex + 1) % len(s.ResponseTimeHistory)
}

// NewServer function to create new servers
func NewServer(serverUrl string, weight, maxConns int) *Server {
	parsedUrl, err := url.Parse(serverUrl)

	if err != nil {
		return nil
	}

	return &Server{
		URL:             parsedUrl,
		Alive:           true,
		Weight:          weight,
		MaxAllowedConns: maxConns,
		LastHealthCheck: time.Now(),
	}
}
