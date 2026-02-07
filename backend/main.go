package main

import (
	"log"
	"net/http"
	"os"

	"skypath/backend/controller"
	"skypath/backend/repository"
	"skypath/backend/service"
)

func main() {
	dataPath := getenv("FLIGHTS_DATA_PATH", "/app/data/flights.json")

	store, err := repository.LoadStore(dataPath)
	if err != nil {
		log.Fatalf("failed to load dataset: %v", err)
	}

	svc := service.NewSearchService(store)
	handler := controller.Router(svc)

	addr := ":" + getenv("PORT", "8080")
	log.Printf("backend listening on %s", addr)

	if err := http.ListenAndServe(addr, handler); err != nil {
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
