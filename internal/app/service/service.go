package service

import (
	"math/rand"
	"net/url"
	"strings"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/saver"
)

type Service struct {
	c    *config.Config
	ds   *saver.Saver
	urls map[string]*config.ShortURL
}

// Constructor
func New(ds *saver.Saver, c *config.Config) *Service {
	s := &Service{}
	s.c = c
	s.ds = ds
	s.urls = make(map[string]*config.ShortURL)
	ds.Load(s.urls)
	return s
}

// Generate and save short url for giver URL
func (s *Service) Post(URL string) (string, error) {
	if len(URL) == 0 {
		return "", config.ErrEmptyReqBody
	}
	if !isURLok(URL) {
		return "", config.ErrURLNotCorrect
	}

	short, isCreated := s.findOrCreateShort(URL)
	if !isCreated {
		return short, nil
	}
	if short == "" {
		return "", config.ErrNoFreeIDs
	}

	newURL := &config.ShortURL{
		URL:   URL,
		Short: short,
	}
	s.urls[newURL.Short] = newURL
	s.ds.Save(newURL)
	return newURL.Short, nil
}

// Get stored URL for giver short url
func (s *Service) Get(ID string) (string, error) {
	recURL, ok := s.urls[ID]
	if !ok {
		return "", config.ErrNoSuchRecord
	}
	return recURL.URL, nil
}

// Generate new short url or return saved for given url,
// bool mean true if Short Url is created, or false if it found.
func (s *Service) findOrCreateShort(url string) (string, bool) {
	for _, rec := range s.urls {
		if strings.EqualFold(url, rec.URL) {
			return rec.Short, false
		}
	}

	rndStr := GetRandStr(s.c.LenShortURL)
	// check: if generated short string for url is already buzy,
	// rerandomize it again. (or change to bigger value types.LenShortUrl)
	const maxTry = 10
	ik := 0
	for _, ok := s.urls[rndStr]; ok && ik < maxTry; {
		ik++
		rndStr = GetRandStr(s.c.LenShortURL)
	}
	if ik == maxTry {
		return "", false
	}

	return rndStr, true
}

func (s *Service) GetLen() int {
	return len(s.urls)
}

func GetRandStr(lenStr int) string {

	var availChars = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	res := make([]byte, lenStr)
	for ik := 0; ik < lenStr; ik++ {
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
