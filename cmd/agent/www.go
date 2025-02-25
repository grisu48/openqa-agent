package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Handler func(http.ResponseWriter, *http.Request)

func createExecHandler() Handler {
	return func(resp http.ResponseWriter, req *http.Request) {
		var job ExecJob
		job.SetDefaults()
		if err := json.NewDecoder(req.Body).Decode(&job); err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(resp, "{\"error\":\"%s\"}", err)
			return
		}

		// Sanity checks
		if err := job.SanityCheck(); err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(resp, "{\"error\":\"%s\"}", err)
			return
		}

		if err := job.exec(); err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(resp, "{\"error\":\"%s\"}", err)
			return
		} else {
			resp.Header().Add("Content-Type", "application/json")
			resp.WriteHeader(200)
			fmt.Fprintf(resp, "{\"status\":\"ok\"}")
		}
	}
}

func createHealthHandler() Handler {
	return func(resp http.ResponseWriter, req *http.Request) {
		resp.Header().Add("Content-Type", "application/json")
		resp.WriteHeader(200)
		fmt.Fprintf(resp, "{\"status\":\"ok\"}")
	}
}
