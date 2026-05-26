package routes

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	_ "github.com/stollenaar/ollamabot/internal/routes/docs"
	httpSwagger "github.com/swaggo/http-swagger"
)

const (
	API = "/api/v1"
)

//	@title			OllamaBot API
//	@version		1.0
//	@description	This is a sample server celler server.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

// @externalDocs.description	OpenAPI
// @externalDocs.url			https://swagger.io/resources/open-api/
func CreateRouter() {
	mux := http.NewServeMux()

	mux.HandleFunc(fmt.Sprintf("GET %s/ping", API), func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"message": "pong"})
	})

	RegisterTradeRoutes(mux)

	// Swagger UI endpoint
	mux.Handle("GET /swagger/", httpSwagger.WrapHandler)
	log.Println("starting server on :8080")
	_ = http.ListenAndServe(":8080", withMiddleware(mux))
}

// withMiddleware adds per-request logging and panic recovery (gin.Default's two
// built-ins). The wrapped ResponseWriter forwards Hijack so the /ws upgrade
// still works.
func withMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		lw := &loggingWriter{ResponseWriter: w, status: 200}
		start := time.Now()
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic %s %s: %v", r.Method, r.URL.Path, rec)
			}
			log.Printf("%d %s %s (%s)", lw.status, r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
		}()
		next.ServeHTTP(lw, r)
	})
}

// loggingWriter records the status code and preserves http.Hijacker (required
// for WebSocket upgrades) and http.Flusher.
type loggingWriter struct {
	http.ResponseWriter
	status int
}

func (w *loggingWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *loggingWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := w.ResponseWriter.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

func (w *loggingWriter) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}

// writeJSON writes v as a JSON response with the given status.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
