package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"backend/db"
)

type userProfileRepository interface {
	GetUserBySubject(ctx context.Context, sub string) (db.User, error)
}

type getUserMeResponse struct {
	ID              int    `json:"id"`
	AuthSubject     string `json:"auth_subject"`
	FirstName       string `json:"first_name"`
	LastName        string `json:"last_name"`
	Email           string `json:"email"`
	ProfileComplete bool   `json:"profile_complete"`
}

func GetUserMe(repo userProfileRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		sub := r.URL.Query().Get("sub")
		if sub == "" {
			writeError(w, http.StatusBadRequest, "sub is required")
			return
		}

		user, err := repo.GetUserBySubject(r.Context(), sub)
		if err != nil {
			if errors.Is(err, db.ErrNotFound) {
				writeError(w, http.StatusNotFound, "user not found")
				return
			}
			writeError(w, http.StatusInternalServerError, "could not fetch user")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(getUserMeResponse{
			ID:              user.ID,
			AuthSubject:     user.AuthSubject,
			FirstName:       user.FirstName,
			LastName:        user.LastName,
			Email:           user.Email,
			ProfileComplete: user.ProfileComplete,
		})
	}
}
