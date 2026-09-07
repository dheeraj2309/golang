package main

import (
	"log"
	"net/http"
)

// a test 	protected route to see the working of middleware

func ProfileHandler(w http.ResponseWriter,r * http.Request){
	username,ok := r.Context().Value(userContextKey).(string)
	if !ok{
		// This should theoretically never happen if the middleware did its job,
		// but it's good Go practice to always check type assertions.
		http.Error(w, "Identity lost in context", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Welcome to your VIP profile, " + username + "!"))
}

func main(){
	store := NewUserStore()
	mx := http.NewServeMux()

	// Register the unprotected routes 
	mx.Handle("/register",LoggerMiddleware(RegisterHandler(store)))
	mx.Handle("/login",LoggerMiddleware(LoginHandler(store)))
	// Register the protected routes
	mx.Handle("GET /profile",LoggerMiddleware(AuthMiddleware(store,http.HandlerFunc(ProfileHandler))))

	addr := "127.0.0.1:8080"
	log.Printf("Server is starting on http://%s...", addr)
	
	// http.ListenAndServe takes the string address and the router.
	err := http.ListenAndServe(addr, mx)
	if err != nil {
		log.Fatal("Server crashed: ", err)
	}
}