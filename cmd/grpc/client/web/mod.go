package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/si-co/vpir-code/cmd/grpc/client/manager"
	"github.com/si-co/vpir-code/lib/client"
	"github.com/si-co/vpir-code/lib/query"
	"github.com/si-co/vpir-code/lib/utils"
	"golang.org/x/xerrors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
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

var staticConfig = true

const keyNotFoundErr string = "no key with the given email id is found"

var staticPointConfig = &utils.Config{
	Servers: map[string]utils.Server{
		"0": {
			IP:   "128.179.33.63",
			Port: 50050,
		},
		"1": {
			IP:   "128.179.33.75",
			Port: 50051,
		},
	},
	Addresses: []string{
		"128.179.33.63:50050", "128.179.33.75:50051",
	},
	ServerCertFile: "/opt/apir/server-cert.pem",
	ServerKeyFile:  "/opt/apir/server-key.pem",
	ClientCertFile: "/opt/apir/client-cert.pem",
	ClientKeyFile:  "/opt/apir/client-key.pem",
}

var staticComplexConfig = &utils.Config{
	Servers: map[string]utils.Server{
		"0": {
			IP:   "128.179.33.63",
			Port: 50040,
		},
		"1": {
			IP:   "128.179.33.75",
			Port: 50041,
		},
	},
	Addresses: []string{
		"128.179.33.63:50040", "128.179.33.75:50041",
	},
	ServerCertFile: "/opt/apir/server-cert.pem",
	ServerKeyFile:  "/opt/apir/server-key.pem",
	ClientCertFile: "/opt/apir/client-cert.pem",
	ClientKeyFile:  "/opt/apir/client-key.pem",
}

var grpcOpts = []grpc.CallOption{
	grpc.UseCompressor(gzip.Name),
	grpc.MaxCallRecvMsgSize(1024 * 1024 * 1024),
	grpc.MaxCallSendMsgSize(1024 * 1024 * 1024),
}

func main() {
	var listenAddr string

	flag.StringVar(&listenAddr, "listen-addr", "", "demo listen address")

	flag.Parse()

	if listenAddr == "" {
		listenAddr = defaultAddr
	}

	pointManager, err := loadPointManager()
	if err != nil {
		log.Fatalf("failed to load point manager: %v", err)
	}

	complexManager, err := loadComplexManager()
	if err != nil {
		log.Fatalf("failed to load complex manager: %v", err)
	}

	pointActor, err := pointManager.Connect()
	if err != nil {
		log.Fatalf("failed to connect point manager: %v", err)
	}

	complexActor, err := complexManager.Connect()
	if err != nil {
		log.Fatalf("failed to connect complex manager: %v", err)
	}

	logger := log.New(os.Stdout, "http: ", log.LstdFlags)

	mux := http.NewServeMux()
	server := &http.Server{
		Handler:  tracing(nextRequestID)(logging(logger)(mux)),
		ErrorLog: logger,
	}

	mux.HandleFunc("/retrieve", gethandleRetreive(pointActor))
	mux.HandleFunc("/count/email", getHandleCountEmail(complexActor))
	mux.HandleFunc("/count/algo", getHandleCountAlgo(complexActor))
	mux.HandleFunc("/count/timestamp", getHandleCountTimestamp(complexActor))

	mux.Handle("/static/", http.FileServer(http.FS(static)))
	mux.HandleFunc("/", handleIndex)

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

func loadPointManager() (manager.Manager, error) {
	configPath := os.Getenv("VPIR_CONFIG_POINT")
	if configPath == "" && !staticConfig {
		return manager.Manager{}, xerrors.New("Please provide " +
			"VPIR_CONFIG_POINT as env variable")
	}

	var config *utils.Config
	var err error

	if staticConfig {
		config = staticPointConfig
	} else {
		config, err = utils.LoadConfig(configPath)
		if err != nil {
			return manager.Manager{}, xerrors.Errorf("failed to load config: %v", err)
		}
	}

	manager := manager.NewManager(*config, grpcOpts)

	return manager, nil
}

func loadComplexManager() (manager.Manager, error) {
	configPath := os.Getenv("VPIR_CONFIG_COMPLEX")
	if configPath == "" && !staticConfig {
		return manager.Manager{}, xerrors.New("Please provide " +
			"VPIR_CONFIG_COMPLEX as env variable")
	}

	var config *utils.Config
	var err error

	if staticConfig {
		config = staticComplexConfig
	} else {
		config, err = utils.LoadConfig(configPath)
		if err != nil {
			return manager.Manager{}, xerrors.Errorf("failed to load config: %v", err)
		}
	}

	manager := manager.NewManager(*config, grpcOpts)

	return manager, nil
}

func handleIndex(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path != "/" {
		http.Error(w, "404 not found", http.StatusNotFound)
		return
	}

	t, err := template.ParseFS(content, "index.html")
	if err != nil {
		http.Error(w, "failed to load template: "+err.Error(),
			http.StatusInternalServerError)
		return
	}

	err = t.Execute(w, nil)
	if err != nil {
		http.Error(w, "failed to execute: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

// POST email=my_email
func gethandleRetreive(actor manager.Actor) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(w, "failed to parse form: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		email := req.PostForm.Get("email")
		if email == "" {
			http.Error(w, "email argument not found", http.StatusBadRequest)
			return
		}

		dbInfo, err := actor.GetDBInfos()
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to get db info: %v", err),
				http.StatusInternalServerError)
			return
		}

		client := client.NewPIR(utils.RandomPRG(), &dbInfo[0])

		result, err := actor.GetKey(email, dbInfo[0], client)
		if err != nil {
			if strings.Contains(err.Error(), keyNotFoundErr) {
				result = "key not found in block"
			} else {
				http.Error(w, fmt.Sprintf("failed to get result: %v", err),
					http.StatusInternalServerError)
				return
			}
		}

		w.Write([]byte(result))
	}
}

// POST position={begin|end}&text="..."
func getHandleCountEmail(actor manager.Actor) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
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

		var fromStart int
		var fromEnd int

		switch position {
		case "begin":
			fromStart = len(text)
			fromEnd = 0
		case "end":
			fromStart = 0
			fromEnd = len(text)
		default:
			http.Error(w, "unknown position: "+position, http.StatusInternalServerError)
			return
		}

		info := &query.Info{
			Target:    query.UserId,
			FromStart: fromStart,
			FromEnd:   fromEnd,
			And:       false,
		}

		clientQuery := info.ToEmailClientFSS(text)

		count, err := executeStatsQuery(clientQuery, actor)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to execute count query: %v", err),
				http.StatusInternalServerError)
			return
		}

		w.Write([]byte(fmt.Sprintf("count: %d", count)))
	}
}

// POST algo={RSA|ed25519|...}
func getHandleCountAlgo(actor manager.Actor) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(w, "failed to parse form: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		algo := req.PostForm.Get("algo")
		if algo == "" {
			http.Error(w, "algo argument not found", http.StatusBadRequest)
			return
		}

		if err != nil {
			http.Error(w, "failed to connect to servers: "+err.Error(),
				http.StatusInternalServerError)
			return
		}

		info := &query.Info{
			Target: query.PubKeyAlgo,
		}

		clientQuery := info.ToPKAClientFSS(algo)

		count, err := executeStatsQuery(clientQuery, actor)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to execute count query: %v", err),
				http.StatusInternalServerError)
		}

		w.Write([]byte(fmt.Sprintf("count: %d", count)))
	}
}

// POST year=N
func getHandleCountTimestamp(actor manager.Actor) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		err := req.ParseForm()
		if err != nil {
			http.Error(w, "failed to parse form: "+err.Error(), http.StatusInternalServerError)
			return
		}

		year := req.PostForm.Get("year")
		if year == "" {
			http.Error(w, "year argument not found", http.StatusBadRequest)
			return
		}

		info := &query.Info{
			Target: query.CreationTime,
		}

		clientQuery := info.ToCreationTimeClientFSS(year)

		count, err := executeStatsQuery(clientQuery, actor)
		if err != nil {
			http.Error(w, fmt.Sprintf("failed to execute count query: %v", err),
				http.StatusInternalServerError)
		}

		w.Write([]byte(fmt.Sprintf("count: %d", count)))
	}
}

// executeStatsQuery takes a client query and executes it
func executeStatsQuery(clientQuery *query.ClientFSS, actor manager.Actor) (uint32, error) {
	in, err := clientQuery.Encode()
	if err != nil {
		return 0, xerrors.Errorf("failed to encode query: %v", err)
	}

	dbInfo, err := actor.GetDBInfos()
	if err != nil {
		return 0, xerrors.Errorf("failed to get db info: %v", err)
	}

	client := client.NewPredicateAPIR(utils.RandomPRG(), &dbInfo[0])

	queries, err := client.QueryBytes(in, len(dbInfo))
	if err != nil {
		return 0, xerrors.Errorf("failed to query bytes: %v", err)
	}

	answers := actor.RunQueries(queries)

	result, err := client.ReconstructBytes(answers)
	if err != nil {
		return 0, xerrors.Errorf("failed to reconstruct bytes: %v", err)
	}

	count, ok := result.(uint32)
	if !ok {
		return 0, xerrors.Errorf("failed to cast result, wrong type %T", result)
	}

	return count, nil
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
