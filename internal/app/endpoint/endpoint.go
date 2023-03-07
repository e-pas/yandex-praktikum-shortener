package endpoint

import (
	"fmt"
	"io"
	"log"
	"net/http"

	. "github.com/e-pas/yandex-praktikum-shortener/internal/app/short_types"
	"github.com/go-chi/chi/v5"
)

type Endpoint struct {
	s Service
}

type Service interface {
	Post(URL string) (*ShortURL, error)
	Get(ID string) (*ShortURL, error)
}

func New(s Service) *Endpoint {
	e := &Endpoint{}
	e.s = s
	return e
}

func (e *Endpoint) Get(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")
	su, err := e.s.Get(urlID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}
	w.Header().Set("Location", su.URL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (e *Endpoint) Post(w http.ResponseWriter, r *http.Request) {

	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err.Error())
		return
	}
	su, err := e.s.Post(string(bodyStr))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}

	shortURL := su.Short
	if ReturnShortWithHost {
		shortURL = OurHost + shortURL
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
	//	url := chi.URLParam(r, "url")
}
