package http

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router, HTTP yönlendirmelerini yönetir
type Router struct {
	chi.Router
}

// NewRouter, yeni bir Router örneği oluşturur
func NewRouter() *Router {
	r := chi.NewRouter()

	// TÜM middleware'leri burada ekleyin - ÖNCE middleware sonra rotalar
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// CORS Headers - hepsini tek bir middleware ile ekleyin
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})

	return &Router{r}
}

// Get, GET isteği için bir handler ekler
func (r *Router) Get(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Get(pattern, handlerFn)
}

// Post, POST isteği için bir handler ekler
func (r *Router) Post(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Post(pattern, handlerFn)
}

// Put, PUT isteği için bir handler ekler
func (r *Router) Put(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Put(pattern, handlerFn)
}

// Delete, DELETE isteği için bir handler ekler
func (r *Router) Delete(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Delete(pattern, handlerFn)
}

// Group, bir yol grubu oluşturur
func (r *Router) Group(fn func(r chi.Router)) {
	r.Router.Group(fn)
}

// Route, bir yol grubu oluşturur
func (r *Router) Route(pattern string, fn func(r chi.Router)) {
	r.Router.Route(pattern, fn)
}

// ServeHTTP, HTTP isteklerini işler
func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	r.Router.ServeHTTP(w, req)
}

// Options, OPTIONS isteği için bir handler ekler
func (r *Router) Options(pattern string, handlerFn http.HandlerFunc) {
	r.Router.Options(pattern, handlerFn)
}

// Use, middleware ekler
func (r *Router) Use(middlewares ...func(http.Handler) http.Handler) {
	r.Router.Use(middlewares...)
}
