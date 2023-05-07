package service

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/url"
	"strings"
	"sync"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/model"
	"github.com/e-pas/yandex-praktikum-shortener/internal/app/repository"
)

type Service struct {
	c    *config.Config
	ds   *repository.Repository
	urls map[string]*model.ShortURL
}

// Constructor
func New(ds *repository.Repository, c *config.Config) *Service {
	s := &Service{}
	s.c = c
	s.ds = ds
	s.urls = make(map[string]*model.ShortURL, 0)
	ds.Load(context.Background(), s.urls)
	return s
}

// Generate and save short url for giver URL
func (s *Service) Post(ctx context.Context, URL string) (string, error) {
	if len(URL) == 0 {
		return "", config.ErrEmptyReqBody
	}
	if !isURLok(URL) {
		return "", config.ErrURLNotCorrect
	}

	userID := ctx.Value(config.ContextKeyUserID).(string)
	short, isCreated := s.findOrCreateShort(URL)
	if isCreated {
		if short == "" {
			return "", config.ErrNoFreeIDs
		}

		newURL := &model.ShortURL{
			URL:     URL,
			Short:   short,
			UserID:  userID,
			Deleted: false,
		}
		if err := s.ds.Save(ctx, *newURL); err != nil {
			return "", err
		}
		s.urls[newURL.Short] = newURL
	}

	if s.c.RetShrtWHost {
		short = s.c.HostName + short
	}

	if !isCreated {
		return short, config.ErrDuplicateURL
	}
	return short, nil
}

func (s *Service) PostBatch(ctx context.Context, URLs []map[string]string) ([]map[string]string, error) {
	if len(URLs) == 0 {
		return nil, config.ErrEmptyReqBody
	}

	userID := ctx.Value(config.ContextKeyUserID).(string)
	hostName := ""
	if s.c.RetShrtWHost {
		hostName = s.c.HostName
	}
	createdURLs := make([]*model.ShortURL, 0) // store slice for new records
	res := make([]map[string]string, 0)       // map with result for browse
	for _, URL := range URLs {
		short, isCreated := s.findOrCreateShort(URL["original_url"])
		if isCreated {
			if short == "" {
				return nil, config.ErrNoFreeIDs
			}

			newURL := &model.ShortURL{
				URL:     URL["original_url"],
				Short:   short,
				UserID:  userID,
				Deleted: false,
			}
			createdURLs = append(createdURLs, newURL)
		}
		rec := make(map[string]string)
		rec["correlation_id"] = URL["correlation_id"]
		rec["short_url"] = hostName + short
		res = append(res, rec)
	}
	err := s.ds.SaveBatch(ctx, createdURLs)
	if err != nil {
		return nil, err
	}
	// merge created records map with  main map
	for _, createRec := range createdURLs {
		s.urls[createRec.Short] = createRec
	}
	return res, nil
}

// Get stored URL for giver short url
func (s *Service) Get(ID string) (string, error) {
	recURL, ok := s.urls[ID]
	if !ok {
		return "", config.ErrNoSuchRecord
	}
	if recURL.Deleted {
		return "", config.ErrURLDeleted
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
	for ik < maxTry {
		if _, ok := s.urls[rndStr]; ok && (ik < maxTry) {
			rndStr = GetRandStr(s.c.LenShortURL)
			ik++
			continue
		}
		break
	}
	if ik == maxTry {
		return "", false
	}

	return rndStr, true
}

// Returns map of short|long urls stored by given user
func (s *Service) GetURLByUser(userID string) []map[string]string {
	res := make([]map[string]string, 0)
	hostName := ""
	if s.c.RetShrtWHost {
		hostName = s.c.HostName
	}
	for _, url := range s.urls {
		if url.UserID == userID {
			rec := make(map[string]string)
			rec["short_url"] = hostName + url.Short
			rec["original_url"] = url.URL
			res = append(res, rec)
		}
	}
	return res
}

func (s *Service) PingDB(ctx context.Context) error {
	return s.ds.Ping(ctx)
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

func markDeleted(ctx context.Context, URL *model.ShortURL, mu *sync.RWMutex) error {
	userID := ctx.Value(config.ContextKeyUserID).(string)
	if userID == URL.UserID {
		mu.Lock()
		defer mu.Unlock()
		if URL.Deleted {
			return fmt.Errorf("link %s already deleted ", URL.Short)
		}
		URL.Deleted = true
	} else {
		return fmt.Errorf("can't delete %s. only owner can ", URL.Short)
	}
	return nil
}

func (s *Service) DeleteURLs(ctx context.Context, shorts []string) error {
	shortURLs := make([]*model.ShortURL, 0)
	for _, short := range shorts {
		url, err := url.Parse(short)
		if err != nil {
			return err
		}
		str := strings.TrimPrefix(url.Path, "/")
		shortURL := s.urls[str]
		shortURLs = append(shortURLs, shortURL)
	}
	go func() {
		proc := NewProcessor(ctx, markDeleted)
		errs := proc.ProceedWith(shortURLs)
		if len(errs) > 0 {
			for _, errv := range errs {
				log.Printf(" error deleting url: %s\n", errv.Error())
			}
		}
		s.ds.UpdateBatch(ctx, shortURLs)
	}()
	return nil
}
