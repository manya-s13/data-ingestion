package main

import (
	"net/http"
	"strings"
)

// authMiddleware validates JWT tokens in request headers
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "Missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}
		
		tokenStr := strings.TrimPrefix(auth, "Bearer ")
		token, err := ValidateJWT(tokenStr)
		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		next(w, r)
	}
}