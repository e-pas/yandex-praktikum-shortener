package endpoint

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/go-chi/chi/v5"
)

type Endpoint struct {
	s servicer
	c *config.Config
}

type servicer interface {
	Post(URL string, userID string) (string, error)
	Get(ID string) (string, error)
	GetURLByUser(userID string) []map[string]string
	PingDB() error
	GetLen() int
}

func New(s servicer, c *config.Config) *Endpoint {
	e := &Endpoint{}
	e.s = s
	e.c = c
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
	userID := r.Context().Value(config.ContextKeyUserID).(string)
	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		log.Print(err.Error())
		return
	}
	shortURL, err := e.s.Post(string(bodyStr), userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(shortURL))
}

func (e *Endpoint) PostAPI(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(config.ContextKeyUserID).(string)
	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	req := map[string]string{}
	err = json.Unmarshal(bodyStr, &req)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}

	shortURL, err := e.s.Post(req[config.PostAPIreqTag], userID)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}

	res := map[string]string{config.PostAPIresTag: shortURL}
	buf, err := json.Marshal(res)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	w.Write(buf)
}

func (e *Endpoint) ShowURLByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(config.ContextKeyUserID).(string)
	urlByUser := e.s.GetURLByUser(userID)
	if len(urlByUser) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	buf, err := json.MarshalIndent(urlByUser, "", "   ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf(" Error: %v", err)))
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(buf)
}

func (e *Endpoint) Ping(w http.ResponseWriter, r *http.Request) {
	err := e.s.PingDB()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("PONG"))
}

func (e *Endpoint) Info(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf(" %d record(s) stored", e.s.GetLen())))
}
