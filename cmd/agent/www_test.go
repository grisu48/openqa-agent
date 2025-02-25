package main

import (
	"context"
	"net/http"
	"testing"
	"time"
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
	if res, err := checkRequest(""); err != nil {
		t.Fatal(err)
	} else if res != http.StatusForbidden {
		t.Fatal("unauthenticated requests are allowed")
	}

	// Check requests with wrong tokens
	if res, err := checkRequest("nots3cr3t"); err != nil {
		t.Fatal(err)
	} else if res != http.StatusForbidden {
		t.Fatal("unauthenticated requests are allowed")
	}

	// Check requests with correct tokens
	if res, err := checkRequest("secret_token"); err != nil {
		t.Fatal(err)
	} else if res != http.StatusAccepted {
		t.Fatal("authenticated requests with token 1 is denied")
	}
	if res, err := checkRequest("secret_token_2"); err != nil {
		t.Fatal(err)
	} else if res != http.StatusAccepted {
		t.Fatal("authenticated requests with token 2 is denied")
	}
}
