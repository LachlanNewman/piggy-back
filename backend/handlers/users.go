package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"reflect"
	"strings"
	"time"

	"backend/db"

	"github.com/go-playground/validator/v10"
)

type userRepository interface {
	CreateUser(ctx context.Context, p db.CreateUserParams) (int, bool, error)
}

type createUserRequest struct {
	AuthSubject string  `json:"auth_subject"  validate:"required"`
	FirstName   string  `json:"first_name"    validate:"required"`
	LastName    string  `json:"last_name"     validate:"required"`
	Email       string  `json:"email"         validate:"required,email"`
	DateOfBirth string  `json:"date_of_birth" validate:"required,datetime=2006-01-02"`
	Weight      float64 `json:"weight"        validate:"gt=0"`
	Gender      string  `json:"gender"        validate:"required,oneof=male female unknown"`
}

type createUserResponse struct {
	ID int `json:"id"`
}

var validate = func() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", 2)[0]
		if name == "" || name == "-" {
			return f.Name
		}
		return name
	})
	return v
}()

func CreateUser(repo userRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var req createUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeError(w, http.StatusBadRequest, "invalid JSON")
			return
		}

		if err := validate.Struct(req); err != nil {
			var ve validator.ValidationErrors
			if errors.As(err, &ve) {
				writeError(w, http.StatusBadRequest, friendlyError(ve[0]))
				return
			}
			writeError(w, http.StatusBadRequest, "invalid request")
			return
		}

		dob, _ := time.Parse("2006-01-02", req.DateOfBirth)

		id, created, err := repo.CreateUser(r.Context(), db.CreateUserParams{
			AuthSubject: req.AuthSubject,
			FirstName:   req.FirstName,
			LastName:    req.LastName,
			Email:       req.Email,
			DateOfBirth: dob,
			Weight:      req.Weight,
			Gender:      req.Gender,
		})
		if err != nil {
			writeError(w, http.StatusInternalServerError, "could not create user")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if created {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusOK)
		}
		json.NewEncoder(w).Encode(createUserResponse{ID: id})
	}
}

func writeError(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func friendlyError(fe validator.FieldError) string {
	field := fe.Field()
	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email address"
	case "gt":
		return field + " must be greater than 0"
	case "oneof":
		return "gender must be one of: male, female, unknown"
	case "datetime":
		return field + " must be in YYYY-MM-DD format"
	default:
		return field + " is invalid"
	}
}
