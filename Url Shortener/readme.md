## URL Shortener
In this project , im making a simple url shortener in  go
I'm not using any database , i will be using a simple map to store the mapping of code -> URL

## Scehmas
* Map with wrapped up lock struct to handle multiple reads and write
* validation schema of what the user is allowed to send on `POST /shorten` -> only URL in its body
* Validation schema of what the user is allowed to send on `GET /{code}` -> only code in the URL to find the org URL

## Work flow of POST
```
Client Request:
POST /shorten
{"url": "https://github.com"}
        │
        ▼
1. Decode & validate request JSON (must start with http:// or https://)
2. Generate 6-char random code (e.g., "k9xL2a")
3. Store in map under Lock: store.urls["k9xL2a"] = "https://github.com"
4. Send 201 Created Response:
   {"code": "k9xL2a", "short_url": "http://localhost:8080/k9xL2a"}
```

## Work flow for GET
```
Client / Browser Request:
GET /k9xL2a
        │
        ▼
1. Extract code: code := r.PathValue("code")
2. Look up under RLock: url, exists := store.urls["k9xL2a"]
3. If not found -> http.Error(w, `{"error": "not found"}`, http.StatusNotFound)
4. If found     -> http.Redirect(w, r, url, http.StatusFound) // 302 Redirect
```

## How random code generated?
Im using the base62 charset to produce the random 6 letter code for the hashing 
Since there is possibility of collision, i have made total 5 reattempts to generate collision free code
```go
const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
func generateRandomCode(length int) (string,error)
```