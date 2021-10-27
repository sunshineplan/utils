package httpsvr

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sunshineplan/utils/cache"
)

var certCache = cache.New(false)

var defaultReload = 24 * time.Hour

// Server defines parameters for running an HTTP server.
type Server struct {
	*http.Server
	Unix string
	Host string
	Port string

	tls      bool
	certFile string
	keyFile  string
	reload   time.Duration
}

// New creates an HTTP server.
func New() *Server {
	return &Server{Server: &http.Server{}}
}

func (s *Server) SetReload(d time.Duration) {
	s.reload = d
}

// Run runs an HTTP server which can be gracefully shut down.
func (s *Server) run(serve func(net.Listener) error) error {
	idleConnsClosed := make(chan struct{})
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit

		if err := s.Shutdown(context.Background()); err != nil {
			log.Println("Failed to close server:", err)
		}
		close(idleConnsClosed)
	}()

	go func() {
		hup := make(chan os.Signal, 1)
		signal.Notify(hup, syscall.SIGHUP)
		for range hup {
			if s.tls {
				cert, err := s.loadCertificate()
				if err != nil {
					log.Println("Failed to reload certificate:", err)
					continue
				}
				if s.reload == 0 {
					s.reload = defaultReload
				}
				certCache.Set("cert", cert, s.reload, s.loadCertificate)
			}
		}
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
		port := s.Port
		if port == "" {
			if s.tls {
				port = "https"
			} else {
				port = "http"
			}
		}
		s.Addr = s.Host + ":" + port
		if err := serve(nil); err != http.ErrServerClosed {
			return fmt.Errorf("failed to server: %v", err)
		}
	}
	<-idleConnsClosed
	return nil
}

func (s *Server) loadCertificate() (interface{}, error) {
	cert, err := tls.LoadX509KeyPair(s.certFile, s.keyFile)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

func (s *Server) getCertificate(_ *tls.ClientHelloInfo) (*tls.Certificate, error) {
	v, ok := certCache.Get("cert")
	if ok {
		return v.(*tls.Certificate), nil
	}

	cert, err := s.loadCertificate()
	if err != nil {
		return nil, err
	}

	if s.reload == 0 {
		s.reload = defaultReload
	}
	certCache.Set("cert", cert, s.reload, s.loadCertificate)

	return cert.(*tls.Certificate), nil
}

// Run runs an HTTP server which can be gracefully shut down.
func (s *Server) Run() error {
	s.tls = false

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
	s.tls = true
	s.certFile = certFile
	s.keyFile = keyFile
	s.TLSConfig = &tls.Config{GetCertificate: s.getCertificate}

	if s.Unix != "" {
		return s.run(func(l net.Listener) error {
			return s.ServeTLS(l, "", "")
		})
	}
	return s.run(func(_ net.Listener) error {
		return s.ListenAndServeTLS("", "")
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
