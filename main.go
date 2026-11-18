package main

import (
	"encoding/json"
	"flag"
	"log"
	"net/http"
	"os"
)

const (
	appPrefix   = "IPINFO_"
	defaultAddr = ":8080"
)

func envOr(env, def string) string {
	if v := os.Getenv(env); v != "" {
		return v
	}
	return def
}

func main() {
	addr := flag.String("addr", envOr(appPrefix+"ADDR", defaultAddr), "Listening address:port")
	asnDB = flag.String("asn", envOr(appPrefix+"DB_ASN", defaultAsnDB), "ASN mmdb file")
	geoDB = flag.String("geoip", envOr(appPrefix+"DB_GEOIP", defaultGeoDB), "GeoIP mmdb file")
	flag.Parse()

	// Load databases
	var err error
	asnReader, err = loadDB(*asnDB)
	if err != nil {
		log.Fatalf("asn: %v", err)
	}
	locReader, err = loadDB(*geoDB)
	if err != nil {
		log.Fatalf("geoip: %v", err)
	}

	// Setup router
	mux := http.NewServeMux()
	mux.HandleFunc("GET /asn/reload", reloadDB)
	mux.HandleFunc("GET /asn/{ip}", asLookup)
	mux.HandleFunc("GET /geo/reload", reloadDB)
	mux.HandleFunc("GET /geo/{ip}", geoLookup)
	mux.HandleFunc("GET /ipinfo/{ip}", ipInfoLookup)

	log.Printf("starting HTTP server on %s ...", *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatal(err)
	}
}

type jsonMessage struct {
	Message string `json:"message,omitempty"`
	Err     error  `json:"error,omitempty"`
}

func renderJSON(w http.ResponseWriter, status int, value any) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(value); err != nil {
		log.Println(err)
	}
}
