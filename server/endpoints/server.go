package endpoints

import (
	"context"
	"crypto/ecdsa"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hashicorp/go-multierror"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/wikisophia/api/server/accounts"
	"github.com/wikisophia/api/server/arguments"
	"github.com/wikisophia/api/server/config"
)

// Server runs the service. Use NewServer() to construct one from an app config,
// and Start() to make it start listening and serving requests.
type Server struct {
	router *httprouter.Router
}

// NewServer makes a server which defines REST endpoints for the service.
func NewServer(key *ecdsa.PrivateKey, store Dependencies) *Server {
	return &Server{
		router: newRouter(key, store),
	}
}

// Dependencies for all the server's endpoints
type Dependencies interface {
	accounts.Emailer
	accounts.Store
	arguments.Store
}

type AccountsStore = accounts.Store
type ArgumentsStore = arguments.Store
type ServerDependencies struct {
	accounts.Emailer
	AccountsStore
	ArgumentsStore
}

func (store ServerDependencies) Close() error {
	var result *multierror.Error
	multierror.Append(result, store.AccountsStore.Close())
	multierror.Append(result, store.ArgumentsStore.Close())
	return result.ErrorOrNil()
}

// Handle exists to make testing easier.
// It lets the Server act without having to bind to a port.
func (s *Server) Handle(w http.ResponseWriter, req *http.Request) {
	s.router.ServeHTTP(w, req)
}

// Start connects the API server to its port and blocks until it hears a
// shutdown signal. Once the server has shut down completely, it adds
// an element to the done channel.
func (s *Server) Start(cfg config.Server, done chan<- struct{}) error {
	var handler http.Handler = s.router
	if len(cfg.CorsAllowedOrigins) > 0 {
		// AllowedMethods should stay in sync with the methods used by the routes
		handler = cors.New(cors.Options{
			AllowedOrigins: cfg.CorsAllowedOrigins,
			AllowedMethods: []string{"DELETE", "GET", "POST", "PATCH"},
			ExposedHeaders: []string{"Location"},
		}).Handler(handler)
	}

	httpServer := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadHeaderTimeout(),
	}

	go shutdownOnSignal(httpServer, done)
	if cfg.UseSSL {
		return httpServer.ListenAndServeTLS(cfg.CertPath, cfg.KeyPath)
	}
	return httpServer.ListenAndServe()
}

func shutdownOnSignal(server *http.Server, done chan<- struct{}) {
	signals := make(chan os.Signal)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)
	sig := <-signals
	log.Printf("Received signal %v. API server shutting down.", sig)
	server.Shutdown(context.Background())
	var s struct{}
	done <- s
}
