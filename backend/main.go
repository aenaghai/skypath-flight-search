package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func main() {
	dataPath := getenv("FLIGHTS_DATA_PATH", "data/flights.json")

	store, err := LoadStore(dataPath)
	if err != nil {
		log.Fatalf("failed to load dataset: %v", err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		withCORS(w, r)
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		withCORS(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			writeJSON(w, http.StatusMethodNotAllowed, SearchError{Message: "method not allowed"})
			return
		}

		origin := r.URL.Query().Get("origin")
		dest := r.URL.Query().Get("destination")
		date := r.URL.Query().Get("date")

		if origin == "" || dest == "" || date == "" {
			writeJSON(w, http.StatusBadRequest, SearchError{Message: "origin, destination, and date are required"})
			return
		}

		itins, err := SearchItineraries(store, origin, dest, date)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, SearchError{Message: err.Error()})
			return
		}

		writeJSON(w, http.StatusOK, SearchResponse{
			Origin:      upper(origin),
			Destination: upper(dest),
			Date:        date,
			Count:       len(itins),
			Itineraries: itins,
		})
	})

	addr := ":" + getenv("PORT", "8080")
	log.Printf("backend listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func getenv(k, def string) string {
	v := os.Getenv(k)
	if v == "" {
		return def
	}
	return v
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func withCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func upper(s string) string {
	n := make([]rune, 0, len(s))
	for _, ch := range s {
		if ch >= 'a' && ch <= 'z' {
			ch = ch - 32
		}
		if ch != ' ' && ch != '\t' && ch != '\n' && ch != '\r' {
			n = append(n, ch)
		}
	}
	return string(n)
}
