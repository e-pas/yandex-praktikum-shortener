package app

import (
	"log"
	"net/http"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/endpoint"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/mware"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/saver"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/service"

	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/chi/v5"
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
	a.r.Use(middleware.RealIP)
	a.r.Use(middleware.Logger)
	a.r.Use(middleware.Recoverer)
	a.r.Use(mware.GzipResponse)
	a.r.Use(mware.GunzipRequest)
	a.r.Use(mware.UserID)

	a.r.Get("/{id}", a.e.Get)
	a.r.Get("/api/user/urls", a.e.ShowURLByUser)
	a.r.Get("/info", a.e.Info)
	a.r.Post("/", a.e.Post)
	a.r.Post("/api/shorten", a.e.PostAPI)
	a.r.Post("/api/shorten/batch", a.e.PostBatchAPI)
	a.r.Get("/ping", a.e.Ping)

	return a, nil
}

func (a *App) Run() error {
	log.Println("service running")
	err := http.ListenAndServe(a.c.Listen, a.r)
	return err
}
