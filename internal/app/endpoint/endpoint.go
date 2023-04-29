package endpoint

import (
	"context"
	"encoding/json"
	"errors"
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
	Post(ctx context.Context, URL string) (string, error)
	PostBatch(ctx context.Context, URLs []map[string]string) ([]map[string]string, error)
	Get(ID string) (string, error)
	GetURLByUser(userID string) []map[string]string
	PingDB(ctx context.Context) error
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
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
		log.Printf(" Error: %v", err)
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
	retStatus := http.StatusCreated
	shortURL, err := e.s.Post(r.Context(), string(bodyStr))
	if err != nil {
		switch {
		default:
			http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
			return
		case errors.Is(err, config.ErrDuplicateURL):
			retStatus = http.StatusConflict
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(retStatus)
	w.Write([]byte(shortURL))
}

func (e *Endpoint) PostAPI(w http.ResponseWriter, r *http.Request) {
	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Body error", http.StatusInternalServerError)
		log.Print(err.Error())
		return
	}

	req := map[string]string{}
	err = json.Unmarshal(bodyStr, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
		return
	}

	retStatus := http.StatusCreated
	shortURL, err := e.s.Post(r.Context(), req[config.PostAPIreqTag])
	if err != nil {
		switch {
		default:
			http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
			return
		case errors.Is(err, config.ErrDuplicateURL):
			retStatus = http.StatusConflict
		}
	}

	res := map[string]string{config.PostAPIresTag: shortURL}
	buf, err := json.Marshal(res)
	if err != nil {
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(retStatus)
	w.Write(buf)
}

func (e *Endpoint) PostBatchAPI(w http.ResponseWriter, r *http.Request) {
	bodyStr, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	req := make([]map[string]string, 0)
	err = json.Unmarshal(bodyStr, &req)
	if err != nil {
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusBadRequest)
		return
	}

	res, err := e.s.PostBatch(r.Context(), req)
	if err != nil {
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusInternalServerError)
		return
	}
	buf, err := json.MarshalIndent(res, "", " ")
	if err != nil {
		http.Error(w, fmt.Sprintf(" Error: %v", err), http.StatusInternalServerError)
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
	err := e.s.PingDB(r.Context())
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
