package app

import (
	"bytes"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"testing"
	"time"

	types "github.com/e-pas/yandex-praktikum-shortener/internal/app/short_types"
	"github.com/stretchr/testify/assert"
)

var pairs []types.ShortURL

const pairnum int = 100

func TestApp_Run(t *testing.T) {
	t.Run("Init test", initTest)
	t.Run("Endpoint POST test", endpointPostTest)
	t.Run("Endpoint GET test", endpointGetTest)
}

func initTest(t *testing.T) {
	tsApp, _ := New()
	go tsApp.Run()
	time.Sleep(500 * time.Millisecond)

	for ik := 0; ik < pairnum; ik++ {
		pairs = append(pairs, types.ShortURL{
			URL: fmt.Sprintf("http://%s.%s", generateRandStr(20), generateRandStr(3)),
		})
	}

}

func endpointPostTest(t *testing.T) {

	for ik := 0; ik < pairnum; ik++ {
		reqBody := prepareBody(pairs[ik].URL)
		resp, err := http.Post("http://localhost:8080", "text/plain", reqBody)
		assert.Nil(t, err)
		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		assert.NotEmpty(t, resp.Body)

		defer resp.Body.Close()
		url, err := io.ReadAll(resp.Body)
		assert.Nil(t, err)
		assert.NotEmpty(t, url)
		pairs[ik].Short = string(url)
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
			retBody:   types.ErrEmptyReqBody.Error()},
		{url: "/",
			body:      "addr.com",
			retStatus: http.StatusBadRequest,
			retBody:   types.ErrURLNotCorrect.Error()},
		{url: "/someurl",
			body:      "",
			retStatus: http.StatusMethodNotAllowed,
			retBody:   ""},
	}

	for _, wt := range wrongTests {
		reqBody := prepareBody(wt.body)
		resp, err := http.Post(fmt.Sprintf("http://localhost:8080%s", wt.url), "text/plain", reqBody)
		assert.Nil(t, err)
		assert.Equal(t, wt.retStatus, resp.StatusCode)
		if len(wt.retBody) > 0 {
			assert.NotEmpty(t, resp.Body)
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
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
		assert.Nil(t, err)
		assert.Equal(t, resp.StatusCode, http.StatusTemporaryRedirect)
		assert.Equal(t, pairs[ik].URL, resp.Header.Get("Location"))
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
			retBody:   types.ErrNoSuchRecord.Error()},
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
			defer resp.Body.Close()
			respBody, err := io.ReadAll(resp.Body)
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

func prepareBody(str string) io.Reader {
	return bytes.NewReader([]byte(str))
}
