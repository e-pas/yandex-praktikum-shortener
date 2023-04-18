package app

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/e-pas/yandex-praktikum-shortener/internal/app/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var pairs []config.ShortURL

const pairnum int = 100

func TestApp_Run(t *testing.T) {
	t.Run("Init test", initTest)
	t.Run("Endpoint POST test", endpointPostTest)
	t.Run("Endpoint POST api test", endpointPostAPITest)
	t.Run("Endpoint GET test", endpointGetTest)
}

func initTest(t *testing.T) {
	tsApp, _ := New()
	go tsApp.Run()
	time.Sleep(500 * time.Millisecond)

	for ik := 0; ik < pairnum; ik++ {
		pairs = append(pairs, config.ShortURL{
			URL: fmt.Sprintf("http://%s.%s", generateRandStr(20), generateRandStr(3)),
		})
	}

}

func endpointPostTest(t *testing.T) {

	for ik := 0; ik < pairnum; ik++ {
		reqBody := prepareBody(pairs[ik].URL, true)
		client := &http.Client{}
		req, err := http.NewRequest(http.MethodPost, "http://localhost:8080", reqBody)
		require.Nil(t, err)
		req.Header.Add("Content-Encoding", "gzip")
		resp, err := client.Do(req)

		require.Nil(t, err)
		require.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		url, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		assert.Nil(t, err)
		assert.NotEmpty(t, url)
		pairs[ik].Short = string(url)
		log.Printf("added rec.%d for url: %s; short key %s", ik, pairs[ik].URL, url)
	}

	wrongTests := []struct {
		url       string
		body      string
		retStatus int
		retBody   string
	}{
		{url: "/",
			body:      "",
			retStatus: http.StatusBadRequest,
			retBody:   config.ErrEmptyReqBody.Error()},
		{url: "/",
			body:      "addr.com",
			retStatus: http.StatusBadRequest,
			retBody:   config.ErrURLNotCorrect.Error()},
		{url: "/someurl",
			body:      "",
			retStatus: http.StatusMethodNotAllowed,
			retBody:   ""},
	}

	for _, wt := range wrongTests {
		reqBody := prepareBody(wt.body, false)
		resp, err := http.Post(fmt.Sprintf("http://localhost:8080%s", wt.url), "text/plain", reqBody)
		require.Nil(t, err)
		assert.Equal(t, wt.retStatus, resp.StatusCode)
		if len(wt.retBody) > 0 {
			assert.NotEmpty(t, resp.Body)
			respBody, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			assert.Nil(t, err)
			assert.Contains(t, string(respBody), wt.retBody)
		}
	}

}

func endpointPostAPITest(t *testing.T) {
	type req struct {
		URL string `json:"url"`
	}
	type res struct {
		Result string `json:"result"`
	}

	for ik := 0; ik < pairnum; ik++ {

		req := req{}
		req.URL = pairs[ik].URL
		reqBody, _ := json.Marshal(req)
		resp, err := http.Post("http://localhost:8080/api/shorten", "text/plain", bytes.NewReader(reqBody))
		require.Nil(t, err)
		require.Equal(t, http.StatusConflict, resp.StatusCode)
		require.Equal(t, resp.Header.Get("Content-Type"), "application/json")
		assert.NotEmpty(t, resp.Body)

		url, err := io.ReadAll(resp.Body)
		defer resp.Body.Close()
		assert.Nil(t, err)
		assert.NotEmpty(t, url)
		res := res{}
		json.Unmarshal(url, &res)
		require.Equal(t, pairs[ik].Short, res.Result)
		log.Printf("checked api rec.%d for url: %s; short key %s", ik, pairs[ik].URL, res.Result)
	}

	wrongTests := []struct {
		url       string
		body      string
		retStatus int
		retBody   string
	}{
		{url: "/",
			body:      "",
			retStatus: http.StatusBadRequest,
			retBody:   config.ErrEmptyReqBody.Error()},
		{url: "/",
			body:      "addr.com",
			retStatus: http.StatusBadRequest,
			retBody:   config.ErrURLNotCorrect.Error()},
		{url: "/someurl",
			body:      "",
			retStatus: http.StatusMethodNotAllowed,
			retBody:   ""},
	}

	for _, wt := range wrongTests {
		reqBody := prepareBody(wt.body, false)
		resp, err := http.Post(fmt.Sprintf("http://localhost:8080%s", wt.url), "text/plain", reqBody)
		require.Nil(t, err)
		assert.Equal(t, wt.retStatus, resp.StatusCode)
		if len(wt.retBody) > 0 {
			assert.NotEmpty(t, resp.Body)
			respBody, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			assert.Nil(t, err)
			assert.Contains(t, string(respBody), wt.retBody)
		}
	}

}

func endpointGetTest(t *testing.T) {
	//	client with disabling redirect
	//	https://golangbyexample.com/http-no-redirect-client-golang/
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	for ik := 0; ik < pairnum; ik++ {
		resp, err := client.Get(pairs[ik].Short)
		require.Nil(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)
		longURL := resp.Header.Get("Location")
		assert.Equal(t, pairs[ik].URL, longURL)
		defer resp.Body.Close()
		log.Printf("checked rec.%d for short key: %s; url: %s, ", ik, pairs[ik].Short, longURL)
	}

	wrongTests := []struct {
		url       string
		retStatus int
		retBody   string
	}{
		{url: "/",
			retStatus: http.StatusMethodNotAllowed,
			retBody:   ""},
		{url: "/someurl",
			retStatus: http.StatusBadRequest,
			retBody:   config.ErrNoSuchRecord.Error()},
		{url: "/someurl/someurl",
			retStatus: http.StatusNotFound,
			retBody:   ""},
	}

	for _, wt := range wrongTests {
		resp, err := client.Get(fmt.Sprintf("http://localhost:8080%s", wt.url))
		assert.Nil(t, err)
		assert.Equal(t, wt.retStatus, resp.StatusCode)
		if len(wt.retBody) > 0 {
			assert.NotEmpty(t, resp.Body)
			respBody, err := io.ReadAll(resp.Body)
			defer resp.Body.Close()
			assert.Nil(t, err)
			assert.Contains(t, string(respBody), wt.retBody)
		}
	}
}

func generateRandStr(l int) string {
	var availChars = []byte("abcdefghijklmnopqrstuvwxyz")

	res := make([]byte, l)
	for ik := 0; ik < l; ik++ {
		res[ik] = availChars[rand.Intn(len(availChars))]
	}

	return string(res)
}

func prepareBody(str string, compr bool) io.Reader {
	res := []byte(str)
	if compr {
		res, _ = Compress(res)
	}
	return bytes.NewReader(res)
}

func Compress(data []byte) ([]byte, error) {
	var b bytes.Buffer
	w := gzip.NewWriter(&b)
	_, err := w.Write(data)
	if err != nil {
		return nil, err
	}
	err = w.Close()
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}
