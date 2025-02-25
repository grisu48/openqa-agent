package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// checkToken checks the given request for a valid authentication token. If not present it rejects the request.
func checkTokenHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, token := range r.Header["Token"] {
			if config.CheckToken(token) {
				next.ServeHTTP(w, r)
				return
			}
		}
		// Deny request
		w.WriteHeader(http.StatusForbidden)
		fmt.Fprintf(w, "{\"error\":\"denied\"}")
	})
}

func execHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var job ExecJob
		job.SetDefaults()
		if err := json.NewDecoder(r.Body).Decode(&job); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}

		// Sanity checks
		if err := job.SanityCheck(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}

		if err := job.exec(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		} else {
			w.Header().Add("Content-Type", "application/json")
			w.WriteHeader(200)
			fmt.Fprintf(w, "{\"status\":\"ok\"}")
		}
	})
}

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "{\"status\":\"ok\"}")
	})
}
