package endpoints

import (
	"crypto/ecdsa"

	"github.com/julienschmidt/httprouter"
)

// newRouter defines the server's API
func newRouter(key *ecdsa.PrivateKey, dependencies Dependencies) *httprouter.Router {
	router := httprouter.New()

	// Accounts
	router.HandlerFunc("POST", "/accounts", accountHandler(dependencies))
	router.POST("/accounts/:id/password", setPasswordHandler(dependencies))
	router.HandlerFunc("POST", "/sessions", postSessionHandler(key, dependencies))

	// Arguments
	router.HandlerFunc("POST", "/arguments", saveHandler(dependencies))
	router.HandlerFunc("GET", "/arguments", getAllArgumentsHandler(dependencies))
	router.GET("/arguments/:id", getLiveArgumentHandler(dependencies))
	router.PATCH("/arguments/:id", updateHandler(dependencies))
	router.DELETE("/arguments/:id", deleteHandler(dependencies))
	router.GET("/arguments/:id/version/:version", getArgumentByVersionHandler(dependencies))

	return router
}
