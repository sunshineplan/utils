package httpsvr

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

// Server defines parameters for running an HTTP server.
type Server struct {
	*http.Server
	Unix string
	Host string
	Port string
}

// New creates an HTTP server. 
func New() *Server {
	return &Server{Server: &http.Server{}}
}

// Run runs an HTTP server which can be gracefully shut down.
func (s *Server) run(serve func(net.Listener) error) error {
	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		if err := s.Shutdown(context.Background()); err != nil {
			fmt.Println("Failed to close server:", err)
		}
		close(idleConnsClosed)
	}()

	if s.Unix != "" {
		listener, err := net.Listen("unix", s.Unix)
		if err != nil {
			return fmt.Errorf("failed to listen socket file: %v", err)
		}
		// Let everyone can access the socket file.
		if err := os.Chmod(s.Unix, 0666); err != nil {
			return fmt.Errorf("failed to chmod socket file: %v", err)
		}
		if err := serve(listener); err != http.ErrServerClosed {
			return fmt.Errorf("failed to server: %v", err)
		}
	} else {
		if s.Host != "" && s.Port != "" {
			s.Addr = s.Host + ":" + s.Port
		}
		if err := serve(nil); err != http.ErrServerClosed {
			return fmt.Errorf("failed to server: %v", err)
		}
	}
	<-idleConnsClosed
	return nil
}

// Run runs an HTTP server which can be gracefully shut down.
func (s *Server) Run() error {
	if s.Unix != "" {
		return s.run(func(l net.Listener) error {
			return s.Serve(l)
		})
	}
	return s.run(func(_ net.Listener) error {
		return s.ListenAndServe()
	})
}

func (s *Server) RunTLS(certFile, keyFile string) error {
	if s.Unix != "" {
		return s.run(func(l net.Listener) error {
			return s.ServeTLS(l, certFile, keyFile)
		})
	}
	return s.run(func(_ net.Listener) error {
		return s.ListenAndServeTLS(certFile, keyFile)
	})
}

// TCP runs an HTTP server on TCP network listener.
func TCP(addr string, handler http.Handler) error {
	return (&Server{Server: &http.Server{Addr: addr, Handler: handler}}).Run()
}

// TLS runs an HTTP server on TCP network listener and handle requests on incoming TLS connections.
func TLS(addr string, handler http.Handler, certFile, keyFile string) error {
	return (&Server{Server: &http.Server{Addr: addr, Handler: handler}}).RunTLS(certFile, keyFile)
}

// Unix runs an HTTP server on Unix domain socket listener.
func Unix(unix string, handler http.Handler) error {
	return (&Server{Unix: unix, Server: &http.Server{Handler: handler}}).Run()
}

// UnixTLS runs an HTTP server on Unix domain socket listener and handle requests on incoming TLS connections.
func UnixTLS(unix string, handler http.Handler, certFile, keyFile string) error {
	return (&Server{Unix: unix, Server: &http.Server{Handler: handler}}).RunTLS(certFile, keyFile)
}
