package controller

import (
	"net/http"

	"skypath/backend/models"
	"skypath/backend/service"
	"skypath/backend/utils"
)

func Router(svc *service.SearchService) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		utils.WithCORS(w, r)
		utils.WriteJSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	mux.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		utils.WithCORS(w, r)

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if r.Method != http.MethodGet {
			utils.WriteJSON(w, http.StatusMethodNotAllowed, models.SearchError{Message: "method not allowed"})
			return
		}

		origin := r.URL.Query().Get("origin")
		dest := r.URL.Query().Get("destination")
		date := r.URL.Query().Get("date")

		if origin == "" || dest == "" || date == "" {
			utils.WriteJSON(w, http.StatusBadRequest, models.SearchError{Message: "origin, destination, and date are required"})
			return
		}

		itins, err := svc.SearchItineraries(origin, dest, date)
		if err != nil {
			utils.WriteJSON(w, http.StatusBadRequest, models.SearchError{Message: err.Error()})
			return
		}

		utils.WriteJSON(w, http.StatusOK, models.SearchResponse{
			Origin:      utils.UpperCompact(origin),
			Destination: utils.UpperCompact(dest),
			Date:        date,
			Count:       len(itins),
			Itineraries: itins,
		})
	})

	return mux
}
