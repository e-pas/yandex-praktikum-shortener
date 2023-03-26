package config

import (
	"errors"
	"log"
	"strings"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Listen       string `env:"LISTEN" envDefault:":8080"`
	HostName     string `env:"HOSTNAME" envDefault:"http://localhost:8080"`
	LenShortURL  int    `env:"SHORTLEN" envDefault:"5"`
	RetShrtWHost bool   `env:"ADDHOST" envDefault:"true"`
}

func New() *Config {
	c := &Config{}
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
	if !strings.HasSuffix(c.HostName, "/") {
		c.HostName = c.HostName + "/"
	}
	return c
}

type ShortURL struct {
	Short string
	URL   string
}

var (
	ErrNoSuchRecord   = errors.New("no such record")
	ErrInvalidReqBody = errors.New("invalid request body")
	ErrEmptyReqBody   = errors.New("empty request body")
	ErrURLNotCorrect  = errors.New("given url is not correct")
	ErrNoFreeIDs      = errors.New("no free short url")
)
