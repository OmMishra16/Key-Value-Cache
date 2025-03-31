package api

import (
    "compress/gzip"
    "encoding/json"
    "net/http"
    "strings"

    "github.com/OmMishra16/key-value-cache/cache"
)

// gzipResponseWriter properly implements http.ResponseWriter
type gzipResponseWriter struct {
    http.ResponseWriter
    writer *gzip.Writer
}

func (w gzipResponseWriter) Write(b []byte) (int, error) {
    return w.writer.Write(b)
}

func (w gzipResponseWriter) Header() http.Header {
    return w.ResponseWriter.Header()
}

func (w gzipResponseWriter) WriteHeader(statusCode int) {
    w.ResponseWriter.WriteHeader(statusCode)
}

// Handler holds handlers for the API
type Handler struct {
    cache *cache.Cache
}

// NewHandler creates a new Handler with the given cache
func NewHandler(cache *cache.Cache) *Handler {
    return &Handler{
        cache: cache,
    }
}

// PutRequest represents the request body for the put operation
type PutRequest struct {
    Key   string `json:"key"`
    Value string `json:"value"`
}

// Response represents the common response structure
type Response struct {
    Status  string `json:"status"`
    Message string `json:"message,omitempty"`
    Key     string `json:"key,omitempty"`
    Value   string `json:"value,omitempty"`
}

// PutHandler handles the PUT operation
func (h *Handler) PutHandler(w http.ResponseWriter, r *http.Request) {
    // Check if method is POST
    if r.Method != http.MethodPost {
        respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    // Parse request body
    var req PutRequest
    decoder := json.NewDecoder(r.Body)
    if err := decoder.Decode(&req); err != nil {
        respondWithError(w, http.StatusBadRequest, "Invalid request body")
        return
    }

    // Validate key and value length
    if len(req.Key) == 0 || len(req.Key) > 256 {
        respondWithError(w, http.StatusBadRequest, "Key length must be between 1 and 256 characters")
        return
    }
    if len(req.Value) > 256 {
        respondWithError(w, http.StatusBadRequest, "Value length must not exceed 256 characters")
        return
    }

    // Put key-value pair in cache
    h.cache.Put(req.Key, req.Value)

    // Respond with success
    response := Response{
        Status:  "OK",
        Message: "Key inserted/updated successfully.",
    }
    respondWithJSON(w, http.StatusOK, response)
}

// GetHandler handles the GET operation
func (h *Handler) GetHandler(w http.ResponseWriter, r *http.Request) {
    // Check if method is GET
    if r.Method != http.MethodGet {
        respondWithError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    // Get key from query parameters
    key := r.URL.Query().Get("key")
    if key == "" {
        respondWithError(w, http.StatusBadRequest, "Key parameter is required")
        return
    }

    // Validate key length
    if len(key) > 256 {
        respondWithError(w, http.StatusBadRequest, "Key length must not exceed 256 characters")
        return
    }

    // Get value from cache
    value, found := h.cache.Get(key)
    if !found {
        response := Response{
            Status:  "ERROR",
            Message: "Key not found.",
        }
        respondWithJSON(w, http.StatusOK, response)
        return
    }

    // Respond with value
    response := Response{
        Status: "OK",
        Key:    key,
        Value:  value,
    }
    respondWithJSON(w, http.StatusOK, response)
}

// Helper function to respond with JSON
func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(statusCode)
    json.NewEncoder(w).Encode(data)
}

// Helper function to respond with an error
func respondWithError(w http.ResponseWriter, statusCode int, message string) {
    response := Response{
        Status:  "ERROR",
        Message: message,
    }
    respondWithJSON(w, statusCode, response)
}

func (h *Handler) Router() http.Handler {
    mux := http.NewServeMux()
    
    // Use separate handlers for different methods
    mux.HandleFunc("/put", h.PutHandler)
    mux.HandleFunc("/get", h.GetHandler)
    
    // Add middleware for common operations
    return h.addMiddleware(mux)
}

func (h *Handler) addMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set common headers first
        w.Header().Set("Content-Type", "application/json")
        w.Header().Set("Connection", "keep-alive")

        // Handle compression if supported
        if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
            gz := gzip.NewWriter(w)
            defer gz.Close()
            w.Header().Set("Content-Encoding", "gzip")
            gzw := &gzipResponseWriter{
                ResponseWriter: w,
                writer:        gz,
            }
            next.ServeHTTP(gzw, r)
            return
        }
        
        next.ServeHTTP(w, r)
    })
}
