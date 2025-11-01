package apperror

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

type appHandler func(w http.ResponseWriter, r *http.Request) error

func JWTMiddleware(jwtSecret string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing token", http.StatusUnauthorized)
			return
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error)  {
			return []byte(jwtSecret), nil
		})
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return 
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			http.Error(w, "Invalid claims", http.StatusUnauthorized)
			return 
		}
		studentID := int64(claims["user_id"].(float64))
		ctx := context.WithValue(r.Context(), "studentID", studentID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func Middleware(h appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var appErr *AppError
		err := h(w, r)

		if err != nil {
			w.Header().Set("Content-Type", "application/json")

			if errors.As(err, &appErr) {

				if errors.Is(err, ErrNotFound) {
					w.WriteHeader(http.StatusNotFound)
					w.Write(ErrNotFound.Marshal())
					return
				}

				err = err.(*AppError)
				w.WriteHeader(http.StatusBadRequest)
				w.Write(appErr.Marshal())
				return 
				
			}

			w.WriteHeader(http.StatusTeapot)
			w.Write(systemError(err).Marshal())
		}
	}
}