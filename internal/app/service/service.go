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

func (s *Service) Post(URL string) (string, error) {
	if len(URL) == 0 {
		return "", types.ErrEmptyReqBody
	}
	if !isURLok(URL) {
		return "", types.ErrURLNotCorrect
	}

	newURL := &types.ShortURL{
		Short: GetRandStr(),
		URL:   URL,
	}
	// check: if same id for url already exists, rerandomize it again.
	const maxTry = 10
	ik := 0
	for _, ok := s.urls[newURL.Short]; ok && ik < maxTry; {
		ik++
		newURL.Short = GetRandStr()
	}
	if ik == maxTry {
		return "", types.ErrNoFreeIDs
	}

	s.urls[newURL.Short] = newURL
	return newURL.Short, nil
}

func (s *Service) Get(ID string) (string, error) {
	recURL, ok := s.urls[ID]
	if !ok {
		return "", types.ErrNoSuchRecord
	}
	return recURL.URL, nil
}

func GetRandStr() string {

	var availChars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	res := make([]byte, types.LenShortURL)
	for ik := 0; ik < types.LenShortURL; ik++ {
		res[ik] = availChars[rand.Intn(len(availChars))]
	}

	return string(res)
}

func isURLok(URL string) bool {
	u, err := url.Parse(URL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return false
	}
	return true
}
