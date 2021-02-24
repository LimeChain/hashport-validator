package router

import (
	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/render"
	"github.com/rs/cors"
	log "github.com/sirupsen/logrus"
	"net/http"
)

const (
	apiV1 = "/api/v1"
)

type APIRouter struct {
	Router *chi.Mux
}

func NewAPIRouter() *APIRouter {
	router := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins: []string{"*"},
	})

	router.Use(
		render.SetContentType(render.ContentTypeJSON),
		middleware.AllowContentType("application/json"),
		middleware.RequestLogger(&middleware.DefaultLogFormatter{Logger: log.StandardLogger()}),
		middleware.RedirectSlashes,
		middleware.Recoverer,
		middleware.NoCache,
		c.Handler)

	return &APIRouter{
		Router: router,
	}
}

func (api *APIRouter) AddV1Router(router http.Handler) {
	api.Router.Mount(apiV1, router)
}
