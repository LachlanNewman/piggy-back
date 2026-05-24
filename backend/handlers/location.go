package handlers

import (
	"context"
	"encoding/json"
	"net/http"
)

type locationRepository interface {
	UpsertUserLocation(ctx context.Context, sub string, lat, lng float64) error
}

type pushLocationBody struct {
	Lat *float64 `json:"lat"`
	Lng *float64 `json:"lng"`
}

func PushLocation(repo locationRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			writeError(w, http.StatusBadRequest, "sub is required")
			return
		}

		var body pushLocationBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if body.Lat == nil {
			writeError(w, http.StatusBadRequest, "lat is required")
			return
		}
		if body.Lng == nil {
			writeError(w, http.StatusBadRequest, "lng is required")
			return
		}

		if err := repo.UpsertUserLocation(r.Context(), sub, *body.Lat, *body.Lng); err != nil {
			writeError(w, http.StatusInternalServerError, "could not update location")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
