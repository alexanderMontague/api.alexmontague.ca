package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"api.alexmontague.ca/helpers"
)

type contextKey string

const UserContextKey contextKey = "user"

type AuthUser struct {
	UserID int
	Email  string
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Missing authorization header"})
			return
		}

		tokenString := strings.Replace(authHeader, "Bearer ", "", 1)
		claims, err := helpers.ValidateToken(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(helpers.Response{Error: true, Code: 401, Message: "Invalid or expired token"})
			return
		}

		authUser := &AuthUser{
			UserID: claims.UserID,
			Email:  claims.Email,
		}

		ctx := context.WithValue(r.Context(), UserContextKey, authUser)
		r = r.WithContext(ctx)

		LogRequest(r)

		next.ServeHTTP(w, r)
	})
}

func GetAuthUser(r *http.Request) *AuthUser {
	user, ok := r.Context().Value(UserContextKey).(*AuthUser)
	if !ok {
		return nil
	}
	return user
}
