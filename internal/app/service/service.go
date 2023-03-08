package service

import (
	"math/rand"
	"net/url"

	types "github.com/e-pas/yandex-praktikum-shortener/internal/app/short_types"
)

type Service struct {
	urls map[string]*types.ShortURL
}

func New() *Service {
	s := &Service{}
	s.urls = make(map[string]*types.ShortURL)
	return s
}

func (s *Service) Post(URL string) (*types.ShortURL, error) {
	if len(URL) == 0 {
		return nil, types.ErrEmptyReqBody
	}
	if !s.isURLok(URL) {
		return nil, types.ErrURLNotCorrect
	}

	newURL := &types.ShortURL{
		Short: s.GetRandStr(),
		URL:   URL,
	}
	// check: if same short url already exists, rerandomize it again.
	for _, ok := s.urls[newURL.Short]; ok; {
		newURL.Short = s.GetRandStr()
	}

	s.urls[newURL.Short] = newURL
	return newURL, nil
}

func (s *Service) Get(ID string) (*types.ShortURL, error) {
	recURL, ok := s.urls[ID]
	if !ok {
		return nil, types.ErrNoSuchRecord
	}
	return recURL, nil
}

func (s *Service) GetRandStr() string {

	var availChars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	res := make([]byte, types.LenShortURL)
	for ik := 0; ik < types.LenShortURL; ik++ {
		res[ik] = availChars[rand.Intn(len(availChars))]
	}

	return string(res)
}

func (s *Service) isURLok(URL string) bool {
	u, err := url.Parse(URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}
