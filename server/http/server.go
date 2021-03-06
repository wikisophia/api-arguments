package http

import (
	"context"
	"crypto/ecdsa"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/wikisophia/api/server/accounts"
	"github.com/wikisophia/api/server/accounts/email"
	accountsHttp "github.com/wikisophia/api/server/accounts/http"
	"github.com/wikisophia/api/server/arguments"
	argumentsHttp "github.com/wikisophia/api/server/arguments/http"
	"github.com/wikisophia/api/server/config"
)

// Server runs the service. Use NewServer() to construct one from an app config,
// and Start() to make it start listening and serving requests.
type Server struct {
	router *httprouter.Router
}

// NewServer makes a server which defines REST endpoints for the service.
func NewServer(key *ecdsa.PrivateKey, store Dependencies) *Server {
	router := httprouter.New()
	accountsHttp.AppendRoutes(router, key, store)
	argumentsHttp.AppendRoutes(router, store)
	return &Server{
		router: router,
	}
}

// Dependencies for all the server's endpoints
type Dependencies interface {
	email.Emailer
	accounts.Store
	arguments.Store
}

type AccountsStore = accounts.Store
type ArgumentsStore = arguments.Store
type ServerDependencies struct {
	email.Emailer
	AccountsStore
	ArgumentsStore
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
