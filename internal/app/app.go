package app

import (
	"fmt"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/endpoint"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/service"
)

type App struct {
	s *service.Service
	e *endpoint.Endpoint
}

func New() (*App, error) {
	a := &App{}

	a.s = service.New()
	a.e = endpoint.New()

	return a, nil
}

func (a *App) Run() error {
	fmt.Println("service running")

	return nil
}
