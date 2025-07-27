/*
*
purpose: manage collection of backend servers and routing logic.
*/
package main

import (
	"net"
	"net/url"
	"strings"
	"sync"
)

type ServerPool struct {
	Servers     []*Server
	CurrentIdx  int
	TotalWeight int
	mutex       sync.RWMutex
}

// AddServer add a new backend server
func (sp *ServerPool) AddServer(server *Server) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	// check for existence
	if _, serv := sp.FindServer(server.URL); serv != nil {
		return
	}

	sp.Servers = append(sp.Servers, server)
	sp.TotalWeight += server.Weight
}

// RemoveServer remove server by url
func (sp *ServerPool) RemoveServer(targetUrl *url.URL) bool {
	serverCount := sp.GetServerCount()

	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	idx, server := sp.FindServer(targetUrl)
	if server != nil {
		sp.Servers = append(sp.Servers[:idx], sp.Servers[idx+1:]...)

		// reset current index if needed
		if idx >= serverCount && serverCount > 0 {
			sp.CurrentIdx = 0
		}
		return true
	}

	return false
}

// GetNextServer main routing logic
func (sp *ServerPool) GetNextServer() *Server {
	serverCount := sp.GetServerCount()

	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	for idx := 0; idx < serverCount; idx++ {
		index := (sp.CurrentIdx + idx) % serverCount

		if sp.Servers[index].CanAcceptConnections() {
			sp.CurrentIdx = (index + 1) % serverCount
			return sp.Servers[index]
		}
	}

	return nil
}

// GetHealthyServers return only alive servers
func (sp *ServerPool) GetHealthyServers() []*Server {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	var healthy []*Server

	for _, server := range sp.Servers {
		if server.CanAcceptConnections() {
			healthy = append(healthy, server)
		}
	}

	// no healthy servers
	if len(healthy) == 0 {
		return nil
	}

	return healthy
}

// GetServerCount return total server count
func (sp *ServerPool) GetServerCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	return len(sp.Servers)
}

// GetAliveServerCount return total healthy server count
func (sp *ServerPool) GetAliveServerCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	return len(sp.GetHealthyServers())
}

// MarkServerDown mark specific server as unhealthy
func (sp *ServerPool) MarkServerDown(targetUrl *url.URL) bool {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	if _, server := sp.FindServer(targetUrl); server != nil {
		server.mutex.Lock()
		server.SetAlive(false)
		server.mutex.Unlock()
		return true
	}

	return false
}

// MarkServerUp mark specific server as healthy
func (sp *ServerPool) MarkServerUp(targetUrl *url.URL) bool {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	if _, server := sp.FindServer(targetUrl); server != nil {
		server.mutex.Lock()
		server.SetAlive(true)
		server.mutex.Unlock()
		return true
	}

	return false
}

// FindServer finds server by URL (assumes mutex is already locked)
func (sp *ServerPool) FindServer(targetUrl *url.URL) (int, *Server) {
	normalizedUrl := NormalizeUrl(targetUrl)

	for idx, server := range sp.Servers {
		if NormalizeUrl(server.URL) == normalizedUrl {
			return idx, server
		}
	}
	return -1, nil
}

// NormalizeUrl returns consistent string for URL comparison
func NormalizeUrl(targetUrl *url.URL) string {
	normalized := *targetUrl // copy to avoid mutating original

	// lowercase host and scheme
	normalized.Host = strings.ToLower(normalized.Host)
	normalized.Scheme = strings.ToLower(normalized.Scheme)

	// remove default ports
	host, port, err := net.SplitHostPort(normalized.Host)
	if err == nil {
		if (normalized.Scheme == "http" && port == "80") ||
			(normalized.Scheme == "https" && port == "443") {
			normalized.Host = host
		}
	}

	// remove trailing slash from path
	if normalized.Path != "/" && normalized.Path != "" {
		normalized.Path = strings.TrimRight(normalized.Path, "/")
	}

	return normalized.String()
}
