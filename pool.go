/*
*
this manages a collection of backend servers and routing logic.
*/
package main

import (
	"net"
	"net/url"
	"strings"
	"sync"
)

type ServerPool struct {
	servers     []*Server
	currentIdx  int          // for round-robin
	totalWeight int          // for weighted routing
	mutex       sync.RWMutex // for thread safety
}

// add new backend server
func (sp *ServerPool) addServer(server *Server) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	sp.servers = append(sp.servers, server)
	sp.totalWeight += server.weight
}

// remove server by URL
func (sp *ServerPool) removeServer(targetUrl *url.URL) {
	sp.mutex.Lock()
	defer sp.mutex.Unlock()

	idx, server := sp.findServer(targetUrl)
	if server != nil {
		sp.servers = append(sp.servers[:idx], sp.servers[idx+1:]...)
		sp.totalWeight -= server.weight
	}
}

// main routing logic (round-robin or weighted)
//func (sp *ServerPool) getNextServer() *Server { //todo
//
//}

// return only alive servers
func (sp *ServerPool) getHealthyServers() []*Server {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	var healthy []*Server

	for _, server := range sp.servers {
		if server.isAlive() {
			healthy = append(healthy, server)
		}
	}

	return healthy
}

// total servers
func (sp *ServerPool) getServerCount() int {
	sp.mutex.RLock()
	defer sp.mutex.RUnlock()

	return len(sp.servers)
}

// healthy servers count
func (sp *ServerPool) getAliveServerCount() int {
	return len(sp.getHealthyServers())
}

// mark specific server as unhealthy
func (sp *ServerPool) markServerDown(targetUrl *url.URL) {
	sp.mutex.Lock()
	_, server := sp.findServer(targetUrl)
	sp.mutex.Unlock()

	if server != nil {
		server.mutex.Lock()
		server.setAlive(false)
		server.mutex.Unlock()
	}
}

// mark specific server as healthy
func (sp *ServerPool) markServerUp(targetUrl *url.URL) {
	sp.mutex.Lock()
	_, server := sp.findServer(targetUrl)
	sp.mutex.Unlock()

	if server != nil {
		server.mutex.Lock()
		server.setAlive(true)
		server.mutex.Unlock()
	}
}

// assuming the sp.mutex is already locked
func (sp *ServerPool) findServer(targetUrl *url.URL) (int, *Server) {
	normalizedUrl := normalizeURL(targetUrl)

	for idx, server := range sp.servers {
		if normalizeURL(server.URL) == normalizedUrl {
			return idx, server
		}
	}

	return -1, nil
}

// normalizeURL returns a consistent string key for URL comparison
func normalizeURL(url *url.URL) string {
	normalized := *url // copy to avoid mutating original

	// lowercase host and scheme
	normalized.Host = strings.ToLower(normalized.Host)
	normalized.Scheme = strings.ToLower(normalized.Scheme)

	// remove default ports
	host, port, err := net.SplitHostPort(normalized.Host)
	if err == nil {
		if (normalized.Scheme == "http" && port == "80") || (normalized.Scheme == "https" && port == "443") {
			normalized.Host = host
		}
	}

	// remove trailing slash from path
	if normalized.Path != "/" {
		normalized.Path = strings.TrimRight(normalized.Path, "/")
	}

	return normalized.String()
}
