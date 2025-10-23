package httpsvr

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunshineplan/utils/cache"
	"github.com/sunshineplan/utils/counter"
	"github.com/sunshineplan/utils/log"
)

var certCache = cache.NewWithRenew[string, *tls.Certificate](false)

var defaultReload = 24 * time.Hour

// Server defines parameters for running an HTTP or HTTPS server.
type Server struct {
	*http.Server
	*log.Logger
	Unix string
	Host string
	Port string

	tls      bool
	certFile string
	keyFile  string
	reload   time.Duration

	l *counter.Listener
}

// New creates a new Server instance with default logger and error log.
func New() *Server {
	logger := log.Default()
	return &Server{Server: &http.Server{ErrorLog: logger.Logger}, Logger: logger}
}

// SetLogger sets a custom logger for both the Server and its internal http.Server.
func (s *Server) SetLogger(logger *log.Logger) {
	s.Logger = logger
	s.Server.ErrorLog = logger.Logger
}

// SetReload defines the certificate reload interval.
// Default is 24 hours if not set explicitly.
func (s *Server) SetReload(d time.Duration) {
	s.reload = d
}

// Serve starts the HTTP or HTTPS server and handles graceful shutdown signals.
//
// It listens on either a Unix domain socket (if s.Unix is set) or a TCP address.
// When receiving SIGHUP, the server reloads configuration or certificates.
// When receiving SIGINT/SIGTERM, it gracefully shuts down all connections.
func (s *Server) Serve(tls bool) (err error) {
	s.tls = tls
	if s.reload == 0 {
		s.reload = defaultReload
	}
	// Channel used to wait for graceful shutdown completion.
	idleConnsClosed := make(chan struct{})
	// Handle system signals for reload and graceful stop.
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
	defer signal.Stop(c)
	go func() {
		for {
			switch <-c {
			case syscall.SIGHUP:
				if err := s.Reload(); err != nil {
					s.Printf("reload failed: %v", err)
				} else {
					s.Print("reload successful")
				}
			case syscall.SIGINT, syscall.SIGTERM:
				ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
				defer cancel()
				if err := s.Shutdown(ctx); err != nil {
					s.Printf("failed to close server: %v", err)
				}
				close(idleConnsClosed)
				return
			}
		}
	}()
	var listener net.Listener
	if s.Unix != "" {
		// Listen on Unix domain socket.
		listener, err = net.Listen("unix", s.Unix)
		if err != nil {
			return fmt.Errorf("failed to listen socket file: %w", err)
		}
		defer os.Remove(s.Unix)
		// Let everyone can access the socket file.
		if err := os.Chmod(s.Unix, 0666); err != nil {
			return fmt.Errorf("failed to chmod socket file: %w", err)
		}
	} else {
		// Default to "http" or "https" if port not specified.
		port := s.Port
		if port == "" {
			if tls {
				port = "https"
			} else {
				port = "http"
			}
		}
		s.Addr = s.Host + ":" + port
		listener, err = net.Listen("tcp", s.Addr)
		if err != nil {
			return fmt.Errorf("failed to listen tcp: %w", err)
		}
	}
	s.l = counter.NewListener(listener)

	if tls {
		err = s.Server.ServeTLS(s.l, "", "")
	} else {
		err = s.Server.Serve(s.l)
	}
	if err != http.ErrServerClosed {
		return fmt.Errorf("failed to serve: %w", err)
	}
	<-idleConnsClosed
	return nil
}

// loadCertificate loads the current TLS certificate and key pair from disk.
func (s *Server) loadCertificate() (*tls.Certificate, error) {
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// getCertificate is a callback for tls.Config.GetCertificate.
// It retrieves the cached certificate, or reloads it if expired.
func (s *Server) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	key := s.certFile + s.keyFile
	v, ok := certCache.Get(key)
	if ok {
		return v, nil
	}
	cert, err := s.loadCertificate()
	if err != nil {
		return nil, err
	}
	certCache.Set(key, cert, s.reload, s.loadCertificate)
	return cert, nil
}

// Run starts an HTTP server with graceful shutdown support.
func (s *Server) Run() error {
	return s.Serve(false)
}

// RunTLS starts an HTTPS server using the provided certificate and key files.
// Certificates are automatically reloaded based on the reload interval.
func (s *Server) RunTLS(certFile, keyFile string) error {
	s.certFile = certFile
	s.keyFile = keyFile
	if s.TLSConfig == nil {
		s.TLSConfig = &tls.Config{}
	}
	s.TLSConfig.GetCertificate = s.getCertificate
	return s.Serve(true)
}

// Reload rotates server's log and reloads TLS certificates if applicable.
func (s *Server) Reload() error {
	s.Rotate()
	if s.tls {
		cert, err := s.loadCertificate()
		if err != nil {
			return fmt.Errorf("failed to reload certificate: %w", err)
		}
		certCache.Set(s.certFile+s.keyFile, cert, s.reload, s.loadCertificate)
	}
	return nil
}

// ReadBytes returns the total number of bytes read by the listener.
func (s *Server) ReadBytes() int64 {
	if s.l == nil {
		return 0
	}
	return s.l.ReadBytes()
}

// WriteBytes returns the total number of bytes written by the listener.
func (s *Server) WriteBytes() int64 {
	if s.l == nil {
		return 0
	}
	return s.l.WriteBytes()
}

// TCP runs an HTTP server using TCP network listener.
func TCP(addr string, handler http.Handler) error {
	return (&Server{Server: &http.Server{Addr: addr, Handler: handler}}).Run()
}

// TLS runs an HTTPS server using TCP network listener.
func TLS(addr string, handler http.Handler, certFile, keyFile string) error {
	return (&Server{Server: &http.Server{Addr: addr, Handler: handler}}).RunTLS(certFile, keyFile)
}

// Unix runs an HTTP server using Unix domain socket listener.
func Unix(unix string, handler http.Handler) error {
	return (&Server{Unix: unix, Server: &http.Server{Handler: handler}}).Run()
}

// UnixTLS runs an HTTPS server using Unix domain socket listener.
func UnixTLS(unix string, handler http.Handler, certFile, keyFile string) error {
	return (&Server{Unix: unix, Server: &http.Server{Handler: handler}}).RunTLS(certFile, keyFile)
}
