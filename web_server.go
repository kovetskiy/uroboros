package main

import (
	"net"
	"net/http"

	"github.com/kovetskiy/lorg"
	"github.com/seletskiy/hierr"
)

const (
	pathStatic                     = "/static/"
	pathAPI                        = "/api/v1/"
	pathBadge                      = "/badge/"
	pathStatus                     = "/status/"
	pathStaticBadges               = "/static/badges/"
	pathStaticBadgeBuildPassing    = "/static/badges/build-passing.svg"
	pathStaticBadgeBuildFailure    = "/static/badges/build-failure.svg"
	pathStaticBadgeBuildProcessing = "/static/badges/build-processing.svg"
)

type WebServer struct {
	mux       *http.ServeMux
	address   string
	logger    *lorg.Log
	resources *resources
	requests  int64
}

func NewWebServer(
	logger *lorg.Log,
	resources *resources,
) *WebServer {
	server := &WebServer{
		mux:       http.NewServeMux(),
		logger:    logger,
		resources: resources,
	}

	server.mux.HandleFunc(
		pathAPI,
		server.HandleAPI,
	)

	server.mux.HandleFunc(
		pathStatic,
		http.FileServer(http.Dir(".")).ServeHTTP,
	)

	server.mux.HandleFunc("/", server.HandleWeb)

	return server
}

func (server *WebServer) Serve(address string) error {
	addr, err := net.ResolveTCPAddr("tcp", address)
	if err != nil {
		return hierr.Errorf(
			err, "can't resolve '%s'", address,
		)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return hierr.Errorf(
			err,
			"can't listen '%s'", addr,
		)
	}

	server.logger.Infof("listening at %s", address)

	return http.Serve(listener, server.mux)
}
