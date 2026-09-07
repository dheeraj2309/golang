package main

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"
)

// we define contextkey new struct which is at core is string but not exactly string
type contextKey string

const userContextKey contextKey = "username"

func AuthMiddleware(s * UserStore,next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter,r * http.Request){
		// Extract the authentication header
		authHeader := r.Header.Get("Authorization")
		if authHeader ==  ""{
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}
		// now we extract the bearer token
		parts := strings.Split(authHeader," ") //Expected format Bearer <token>
		if len(parts) != 2 || parts[0] != "Bearer"{
			http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
			return
		}
		token := parts[1]

		const prefix = "secret-token-for-"
		if !strings.HasPrefix(token,prefix){
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract the username from our mock token
		username := strings.TrimPrefix(token, prefix)

		_, err := s.GetUser(username)
		if err != nil {
			// If GetUser returns an error (like ErrUserNotFound), we kick them out.
			http.Error(w, "User does not exist or token is invalid", http.StatusUnauthorized)
			return
		}
		// Create a new context with the verified username
		ctx := context.WithValue(r.Context(), userContextKey, username)

		// Clone the request with the new context
		rWithCtx := r.WithContext(ctx)

		// Pass control to the next handler in the chain, using the cloned request
		next.ServeHTTP(w, rWithCtx)
	})
	
}

// To store the status codes of the requests , we extend(kind of) the basic http.ResponseWriter by attaching the statusCodes with it in an interface
type responseRecoder struct{
	http.ResponseWriter // struct embedding
	statusCode int
}

// we override the normal behaviour of the WriteHeader behavior via this new struct methods
func (rec * responseRecoder) WriteHeader(code int){
	rec.statusCode = code // this declared here so our logger can read this
	rec.ResponseWriter.WriteHeader(code)
}

// now we actually create the loggingMiddleware

func LoggerMiddleware(next http.Handler) http.Handler{
	return http.HandlerFunc(func (w http.ResponseWriter,r * http.Request){
		start := time.Now()
		recorder := &responseRecoder{
			ResponseWriter: w,
			statusCode: http.StatusOK,
		}
		next.ServeHTTP(recorder,r)

		duration := time.Since(start)
		// Log the result, reading the intercepted status code from our Spy
		log.Printf("[%s] %s | Status: %d | Time: %v", 
			r.Method, r.URL.Path, recorder.statusCode, duration)
	})
}