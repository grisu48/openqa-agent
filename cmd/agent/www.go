package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

func getFileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		paths := values["path"]
		if len(paths) <= 0 || paths[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"missing 'path' argument\"}")
			return
		}
		file, err := os.OpenFile(paths[0], os.O_RDONLY, 0600)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				fmt.Fprintf(w, "{\"error\":\"file not found\"}")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		defer file.Close()
		// Get file size
		size, err := file.Seek(0, 2)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		_, err = file.Seek(0, 0)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		w.Header().Add("Content-Length", fmt.Sprintf("%d", size))
		w.Header().Add("Content-Type", "application/octet-stream")
		w.Header().Add("Content-Disposition", "attachment")
		w.WriteHeader(http.StatusAccepted)

		buf := make([]byte, 4096)
		for {
			n, err := file.Read(buf)
			// Always write the data first
			if n > 0 {
				if _, err := w.Write(buf[:n]); err != nil {
					// Assume connection has been closed and don't do anything
					return
				}
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					log.Fatalf("io error while reading '%s': %s", paths[0], err)
					return
				}
			}
		}
	})
}

func putFileHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		values := r.URL.Query()
		paths := values["path"]
		if len(paths) <= 0 || paths[0] == "" {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"missing 'path' argument\"}")
			return
		}

		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "{\"error\":\"missing body\"}")
			return
		}

		// By default create or overwrite a file, and set the permissions to 0644
		var mode os.FileMode = 0644
		flag := os.O_WRONLY | os.O_CREATE | os.O_SYNC
		file, err := os.OpenFile(paths[0], flag, mode)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
			return
		}
		defer file.Close()

		// Write body to file
		buf := make([]byte, 4096)
		var received uint64
		for {
			n, err := r.Body.Read(buf)
			// Always write the data first
			if n > 0 {
				if _, err := file.Write(buf[:n]); err != nil {
					log.Fatalf("io error while writing '%s': %s", paths[0], err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
					return
				}
				received += uint64(n)
			}
			if err != nil {
				if errors.Is(err, io.EOF) {
					break
				} else {
					log.Fatalf("io error while receiving '%s': %s", paths[0], err)
					w.WriteHeader(http.StatusInternalServerError)
					fmt.Fprintf(w, "{\"error\":\"%s\"}", err)
					return
				}
			}
		}
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprintf(w, "{\"status\":\"ok\",\"received\":%d}", received)
	})
}

func healthHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(200)
		fmt.Fprintf(w, "{\"status\":\"ok\"}")
	})
}
