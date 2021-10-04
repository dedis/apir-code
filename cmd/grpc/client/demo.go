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

//go:embed index.html
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

	mux.HandleFunc("/retrieve", lc.handleRetreive)
	mux.HandleFunc("/count/email", lc.handleCountEmail)
	mux.HandleFunc("/count/algo", lc.handleCountAlgo)
	mux.HandleFunc("/count/timestamp", lc.handleCountTimestamp)

	mux.Handle("/static/", http.FileServer(http.FS(static)))
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
	if req.URL.Path != "/" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	t, err := template.ParseFS(content, "index.html")
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

// POST position={begin|end}&text="..."
func (lc *localClient) handleCountEmail(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	position := req.PostForm.Get("position")
	if position == "" {
		http.Error(w, "position argument not found", http.StatusBadRequest)
		return
	}

	text := req.PostForm.Get("text")
	if text == "" {
		http.Error(w, "text argument not found", http.StatusBadRequest)
		return
	}

	err = lc.connectToServers()
	defer lc.closeConnections()

	if err != nil {
		http.Error(w, "failed to connect to servers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lc.flags.target = "email"
	lc.flags.and = false
	lc.flags.id = text

	switch position {
	case "begin":
		lc.flags.fromStart = len(text)
		lc.flags.fromEnd = 0
	case "end":
		lc.flags.fromStart = 0
		lc.flags.fromEnd = len(text)
	default:
		http.Error(w, "unknown position: "+position, http.StatusInternalServerError)
		return
	}

	result, err := lc.exec()
	if err != nil {
		http.Error(w, "failed to get response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(result))
}

// POST algo={rsa|ed25519|...}
func (lc *localClient) handleCountAlgo(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	algo := req.PostForm.Get("algo")
	if algo == "" {
		http.Error(w, "algo argument not found", http.StatusBadRequest)
		return
	}

	err = lc.connectToServers()
	defer lc.closeConnections()

	if err != nil {
		http.Error(w, "failed to connect to servers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lc.flags.target = "algo"
	lc.flags.and = false
	lc.flags.id = algo // ??

	result, err := lc.exec()
	if err != nil {
		http.Error(w, "failed to get response: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Write([]byte(result))
}

// POST day=N&month=N&year=N
func (lc *localClient) handleCountTimestamp(w http.ResponseWriter, req *http.Request) {
	err := req.ParseForm()
	if err != nil {
		http.Error(w, "failed to parse form: "+err.Error(), http.StatusInternalServerError)
		return
	}

	day := req.PostForm.Get("day")
	if day == "" {
		http.Error(w, "day argument not found", http.StatusBadRequest)
		return
	}

	month := req.PostForm.Get("month")
	if month == "" {
		http.Error(w, "month argument not found", http.StatusBadRequest)
		return
	}

	year := req.PostForm.Get("year")
	if year == "" {
		http.Error(w, "year argument not found", http.StatusBadRequest)
		return
	}

	err = lc.connectToServers()
	defer lc.closeConnections()

	if err != nil {
		http.Error(w, "failed to connect to servers: "+err.Error(), http.StatusInternalServerError)
		return
	}

	t, err := time.Parse("2-1-2006", fmt.Sprintf("%s-%s-%s", day, month, year))
	if err != nil {
		http.Error(w, "failed to parse time: "+err.Error(), http.StatusInternalServerError)
		return
	}

	lc.flags.target = "creation"
	lc.flags.and = false
	lc.flags.id = t.String() // ??

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
