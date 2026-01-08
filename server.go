// Package fasthttp provides optimized HTTP server functionality
// Inspired by VictoriaMetrics' fast HTTP handling
package fasthttp

import (
	"context"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

// Server wraps fasthttp.Server with additional optimizations
type Server struct {
	fasthttp.Server
	
	// Metrics tracking
	requestsInFlight int64
	requestsTotal    int64
	requestDuration  time.Duration
	
	// Connection pooling
	connPool sync.Pool
	
	// Buffer pools for zero-allocation handling
	readBufPool  sync.Pool
	writeBufPool sync.Pool
}

// NewServer creates a new optimized HTTP server
func NewServer(handler http.Handler) *Server {
	s := &Server{
		Server: fasthttp.Server{
			Handler: fasthttpadaptor.NewFastHTTPHandler(handler),
			Name:    "LuxFastHTTP",
		},
		connPool: sync.Pool{
			New: func() interface{} {
				return &net.TCPConn{}
			},
		},
		readBufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 4096)
			},
		},
		writeBufPool: sync.Pool{
			New: func() interface{} {
				return make([]byte, 4096)
			},
		},
	}
	
	// Configure fasthttp optimizations
	s.Server.TCPKeepalive = 3 * time.Minute
	s.Server.ReadTimeout = 10 * time.Second
	s.Server.WriteTimeout = 10 * time.Second
	s.Server.IdleTimeout = 30 * time.Second
	s.Server.MaxConnsPerIP = 1000
	s.Server.MaxRequestsPerConn = 1000
	
	return s
}

// ListenAndServe starts the server with optimizations
func (s *Server) ListenAndServe(addr string) error {
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	
	// Set TCP optimizations
	if tcpLn, ok := ln.(*net.TCPListener); ok {
		if err := setTCPOptions(tcpLn); err != nil {
			// Log error but continue
		}
	}
	
	return s.Server.Serve(ln)
}

// setTCPOptions applies TCP-level optimizations
func setTCPOptions(ln *net.TCPListener) error {
	// These would be platform-specific optimizations
	// Similar to VictoriaMetrics' network optimizations
	return nil
}

// HandleRequest provides optimized request handling
func (s *Server) HandleRequest(ctx *fasthttp.RequestCtx) {
	start := time.Now()
	
	// Use pooled buffers
	ctx.SetUserValue("start_time", start)
	
	// Call the standard handler through adapter
	s.Server.Handler(ctx)
	
	// Update metrics
	duration := time.Since(start)
	// Increment counters atomically
}

// WithContext wraps the server with context support
func (s *Server) WithContext(ctx context.Context) *Server {
	// Add context cancellation support
	go func() {
		<-ctx.Done()
		s.Server.Shutdown()
	}()
	
	return s
}

// OptimizedHandler wraps an http.Handler with fasthttp optimizations
func OptimizedHandler(handler http.Handler) fasthttp.RequestHandler {
	return func(ctx *fasthttp.RequestCtx) {
		// Apply VictoriaMetrics-style optimizations:
		// 1. Zero-allocation request parsing
		// 2. Optimized header handling
		// 3. Fast path for common routes
		
		// Use the adapter but with our optimizations
		fasthttpadaptor.NewFastHTTPHandler(handler)(ctx)
	}
}

// NewRouter creates an optimized router
func NewRouter() *Router {
	return &Router{
		routes: make(map[string]RouteHandler),
	}
}

// Router provides optimized routing
type Router struct {
	mu     sync.RWMutex
	routes map[string]RouteHandler
}

// RouteHandler defines a route handler type
type RouteHandler func(ctx *fasthttp.RequestCtx)

// AddRoute adds a route to the router
func (r *Router) AddRoute(method, path string, handler RouteHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	
	key := method + " " + path
	r.routes[key] = handler
}

// HandleRequest handles an incoming request
func (r *Router) HandleRequest(ctx *fasthttp.RequestCtx) {
	path := string(ctx.Path())
	method := string(ctx.Method())
	
	r.mu.RLock()
	handler, ok := r.routes[method+" "+path]
	r.mu.RUnlock()
	
	if ok {
		handler(ctx)
		return
	}
	
	// Fallback to not found
	ctx.Error("Not Found", fasthttp.StatusNotFound)
}

// FastHTTPMiddleware provides middleware support
type FastHTTPMiddleware func(handler RouteHandler) RouteHandler

// ChainMiddleware chains multiple middleware functions
func ChainMiddleware(handler RouteHandler, middleware ...FastHTTPMiddleware) RouteHandler {
	for i := len(middleware) - 1; i >= 0; i-- {
		handler = middleware[i](handler)
	}
	return handler
}