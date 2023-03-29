package app

import (
	"log"
	"net/http"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/endpoint"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/saver"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type App struct {
	c  *config.Config
	ds *saver.Saver
	s  *service.Service
	e  *endpoint.Endpoint
	r  chi.Router
}

func New() (*App, error) {
	a := &App{}
	a.c = config.New()
	a.ds = saver.New(a.c)
	a.s = service.New(a.ds, a.c)
	a.e = endpoint.New(a.s, a.c)
	a.r = chi.NewRouter()
	a.r.Use(middleware.RequestID)
	a.r.Use(middleware.RealIP)
	a.r.Use(middleware.Logger)
	a.r.Use(middleware.Recoverer)
	a.r.Use(config.GzipHandle)

	a.r.Get("/{id}", a.e.Get)
	a.r.Get("/info", a.e.Info)
	a.r.Post("/", a.e.Post)
	a.r.Post("/api/shorten", a.e.PostAPI)

	return a, nil
}

func (a *App) Run() error {
	log.Println("service running")
	err := http.ListenAndServe(a.c.Listen, a.r)
	return err
}
