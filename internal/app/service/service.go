package service

import (
	"math/rand"
	"net/url"
	"strings"

	types "github.com/e-pas/yandex-praktikum-shortener/internal/app/short_types"
)

type Service struct {
	urls map[string]*types.ShortURL
}

// Constructor
func New() *Service {
	s := &Service{}
	s.urls = make(map[string]*types.ShortURL)
	return s
}

// Generate and save short url for giver URL
func (s *Service) Post(URL string) (string, error) {
	if len(URL) == 0 {
		return "", types.ErrEmptyReqBody
	}
	if !isURLok(URL) {
		return "", types.ErrURLNotCorrect
	}

	newURL := &types.ShortURL{
		URL:   URL,
		Short: s.findOrCreateShort(URL),
	}

	if newURL.Short == "" {
		return "", types.ErrNoFreeIDs
	}

	s.urls[newURL.Short] = newURL
	return newURL.Short, nil
}

// Get stored URL for giver short url
func (s *Service) Get(ID string) (string, error) {
	recURL, ok := s.urls[ID]
	if !ok {
		return "", types.ErrNoSuchRecord
	}
	return recURL.URL, nil
}

// Generate new short url or return saved for given url
func (s *Service) findOrCreateShort(url string) string {
	for _, rec := range s.urls {
		if strings.EqualFold(url, rec.URL) {
			return rec.Short
		}
	}

	rndStr := GetRandStr()
	// check: if generated short string for url is already buzy,
	// rerandomize it again. (or change to bigger value types.LenShortUrl)
	const maxTry = 10
	ik := 0
	for _, ok := s.urls[rndStr]; ok && ik < maxTry; {
		ik++
		rndStr = GetRandStr()
	}
	if ik == maxTry {
		return ""
	}

	return rndStr
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
