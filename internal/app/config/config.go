package config

import (
	"errors"
	"flag"
	"log"
	"strings"

	"github.com/caarlos0/env/v7"
)

type Config struct {
	Listen       string `env:"SERVER_ADDRESS"`
	HostName     string `env:"BASE_URL"`
	FileStorage  string `env:"FILE_STORAGE_PATH"`
	LenShortURL  int    `env:"SHORTLEN"`
	RetShrtWHost bool   `env:"ADDHOST" envDefault:"true"`
}

const (
	PostAPIreqTag string = "url"
	PostAPIresTag string = "result"
	CookieName    string = "ShrtnrUserID"
	PassCiph      string = "AF12345"
)

func New() *Config {
	c := &Config{}
	err := env.Parse(c)
	if err != nil {
		log.Fatal(err)
	}
	if c.Listen == "" {
		flag.StringVar(&c.Listen, "a", ":8080", "HTTP listen addr")
	}
	if c.HostName == "" {
		flag.StringVar(&c.HostName, "b", "http://localhost:8080", "Host name in short URL")
	}
	if c.FileStorage == "" {
		flag.StringVar(&c.FileStorage, "f", "", "File to store. If omitted no files will created")
	}
	if c.LenShortURL == 0 {
		flag.IntVar(&c.LenShortURL, "l", 5, "Length of short address")
	}
	flag.Parse()
	if !strings.HasSuffix(c.HostName, "/") {
		c.HostName = c.HostName + "/"
	}
	return c
}

type ShortURL struct {
	Short  string `json:"SHORT"`
	URL    string `json:"URL"`
	UserID string `json:"USERID"`
}

var (
	ErrNoSuchRecord   = errors.New("no such record")
	ErrInvalidReqBody = errors.New("invalid request body")
	ErrEmptyReqBody   = errors.New("empty request body")
	ErrURLNotCorrect  = errors.New("given url is not correct")
	ErrNoFreeIDs      = errors.New("no free short url")
	ErrInvalidGZip    = errors.New("error in gzipped request")
)
