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
	Unix    string
	Host    string
	Port    string
	Handler http.Handler
}

// Run runs an HTTP server which can be gracefully shut down.
func (s *Server) Run() error {
	server := &http.Server{Handler: s.Handler}

	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		if err := server.Shutdown(context.Background()); err != nil {
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
		if err := server.Serve(listener); err != http.ErrServerClosed {
			return fmt.Errorf("failed to server: %v", err)
		}
	} else {
		if s.Host != "" && s.Port != "" {
			server.Addr = s.Host + ":" + s.Port
		}
		if err := server.ListenAndServe(); err != http.ErrServerClosed {
			return fmt.Errorf("failed to server: %v", err)
		}
	}
	<-idleConnsClosed
	return nil
}

// TCP runs an HTTP server on TCP network listener.
func TCP(hostport string, handler http.Handler) error {
	if hostport == "" {
		hostport = ":http"
	}
	host, port, err := net.SplitHostPort(hostport)
	if err != nil {
		return err
	}
	return (&Server{Host: host, Port: port, Handler: handler}).Run()
}

// Unix runs an HTTP server on Unix domain socket listener.
func Unix(unix string, handler http.Handler) error {
	return (&Server{Unix: unix, Handler: handler}).Run()
}
