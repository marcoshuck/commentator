package main

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log"
	"net/http"
)

type Server struct {
	router   http.Handler
	listener *http.Server
	logger   *log.Logger
}

func (s *Server) ListenAndServe() error {
	s.logger.Println("Listening on:", s.listener.Addr)
	return s.listener.ListenAndServe()
}

func NewServer(logger *log.Logger, service Service) Server {
	return newServer(newListener(logger, 8080), newRouter(service), logger)
}

func newRouter(s Service) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Handle("/", s.Routes())
	return r
}

func newListener(logger *log.Logger, port uint) *http.Server {
	return &http.Server{
		ErrorLog: logger,
		Addr:     fmt.Sprintf(":%d", port),
	}
}

func newServer(listener *http.Server, router http.Handler, logger *log.Logger) Server {
	listener.Handler = router
	return Server{
		logger:   logger,
		router:   router,
		listener: listener,
	}
}
