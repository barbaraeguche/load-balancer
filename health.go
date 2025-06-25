/*
* this monitors backend server health and update their status.
 */
package main

import (
	"context"
	"net/http"
	"net/url"
	"time"
)

type HealthChecker struct {
	serverPool *ServerPool
	interval   time.Duration
	timeout    time.Duration
	httpClient *http.Client
	context    context.Context
}

// begin health checking goroutine
func start() {

}

// gracefully stop health checking
func stop() {

}

// check individual server health
func checkServer(server *Server) {

}

// iterate through all servers
func checkAllServers() {

}

// make HTTP request to server
func isServerHealthy(targetUrl *url.URL) {

}
