package main

import (
	"context"
	"embed"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"time"
)

const defaultAddr = ":9990"

type key int

const (
	requestIDKey key = 0
)

//go:embed index.gohtml
var content embed.FS

//go:embed static
var static embed.FS

func (lc *localClient) runDemo() {
	var listenAddr = lc.flags.listenAddr
	if listenAddr == "" {
		listenAddr = defaultAddr
	}

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	mux := http.NewServeMux()
	server := &http.Server{
		Handler:  tracing(nextRequestID)(logging(logger)(mux)),
		ErrorLog: logger,
	}

	mux.HandleFunc("/retreive", lc.handleRetreive)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.FS(static))))
	mux.HandleFunc("/", lc.handleIndex)

	ln, err := net.Listen("tcp", listenAddr)
	if err != nil {
		logger.Fatalf("failed to create conn '%s': %v", listenAddr, err)
		return
	}

	logger.Printf("Server is ready to handle request at %s", ln.Addr())

	err = server.Serve(ln)
	if err != nil && err != http.ErrServerClosed {
		logger.Fatal(err)
	}
}

func (lc *localClient) handleIndex(w http.ResponseWriter, req *http.Request) {
	t, err := template.ParseFS(content, "index.gohtml")
	if err != nil {
		http.Error(w, "failed to load template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "failed to execute: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST email=my_email
func (lc *localClient) handleRetreive(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	email := req.PostForm.Get("email")
	if email == "" {
		http.Error(w, "email argument not found", http.StatusBadRequest)
		return
	}

	err = lc.connectToServers()
	defer lc.closeConnections()

	if err != nil {
		http.Error(w, "failed to connect to servers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lc.flags.id = email

	result, err := lc.exec()
	if err != nil {
		http.Error(w, "failed to get response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(result))
}

func nextRequestID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func logging(logger *log.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				requestID, ok := r.Context().Value(requestIDKey).(string)
				if !ok {
					requestID = "unknown"
				}
				logger.Println(requestID, r.Method, r.URL.Path, r.RemoteAddr, r.UserAgent())
			}()
			next.ServeHTTP(w, r)
		})
	}
}

func tracing(nextRequestID func() string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-Id")
			if requestID == "" {
				requestID = nextRequestID()
			}
			ctx := context.WithValue(r.Context(), requestIDKey, requestID)
			w.Header().Set("X-Request-Id", requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
