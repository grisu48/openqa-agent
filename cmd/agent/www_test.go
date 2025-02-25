package main

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// return 200 and exit
func dummyHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	})
}

func TestTokenHandler(t *testing.T) {
	var cf Config
	cf.SetDefaults()
	cf.Token = append(cf.Token, Token{Token: "secret_token"})
	cf.Token = append(cf.Token, Token{Token: "secret_token_2"})

	sm := http.NewServeMux()
	sm.Handle("/dummy", checkTokenHandler(dummyHandler(), cf))
	server := http.Server{Addr: "127.0.0.1:8421", Handler: sm}
	go func() {
		server.ListenAndServe()
	}()
	defer server.Shutdown(context.Background())
	time.Sleep(1 * time.Second)

	// Request function. Perform a request with the given token (if any) and return the http status code or any error
	checkRequest := func(token string) (int, error) {
		client := &http.Client{}
		req, err := http.NewRequest("GET", "http://127.0.0.1:8421/dummy", nil)
		if err != nil {
			return 0, err
		}
		if token != "" {
			req.Header.Add("Token", token)
		}
		res, err := client.Do(req)
		return res.StatusCode, err
	}

	// Check unauthenticated requests
	res, err := checkRequest("")
	assert.NoError(t, err, "checkRequest should succeed")
	assert.Equal(t, res, http.StatusForbidden, "unauthenticated requests should be forbidden")

	// Check requests with wrong tokens
	res, err = checkRequest("nots3cr3t")
	assert.NoError(t, err, "checkRequest should succeed")
	assert.Equal(t, res, http.StatusForbidden, "requests with wrong token should be forbidden")

	// Check requests with correct tokens
	res, err = checkRequest("secret_token")
	assert.NoError(t, err, "checkRequest should succeed")
	assert.Equal(t, res, http.StatusAccepted, "requests with correct token 1 should succeed")
	res, err = checkRequest("secret_token_2")
	assert.NoError(t, err, "checkRequest should succeed")
	assert.Equal(t, res, http.StatusAccepted, "requests with correct token 2 should succeed")
}
