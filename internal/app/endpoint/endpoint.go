package endpoint

import (
	"fmt"
	"io"
	"log"
	"net/http"

	types "github.com/e-pas/yandex-praktikum-shortener/internal/app/short_types"
	"github.com/go-chi/chi/v5"
)

type Endpoint struct {
	s service
}

type service interface {
	Post(URL string) (string, error)
	Get(ID string) (string, error)
}

func New(s service) *Endpoint {
	e := &Endpoint{}
	e.s = s
	return e
}

func (e *Endpoint) Get(w http.ResponseWriter, r *http.Request) {
	urlID := chi.URLParam(r, "id")
	longURL, err := e.s.Get(urlID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}
	w.Header().Set("Location", longURL)
	w.WriteHeader(http.StatusTemporaryRedirect)
}

func (e *Endpoint) Post(w http.ResponseWriter, r *http.Request) {

	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err.Error())
		return
	}
	shortURL, err := e.s.Post(string(bodyStr))
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}

	if types.ReturnShortWithHost {
		shortURL = types.OurHost + shortURL
	}
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}
