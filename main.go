package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"time"
)

type Product struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Price       int    `json:"price"`
	Currency    string `json:"currency"`
	Emoji       string `json:"emoji"`
	Description string `json:"description,omitempty"`
	Stock       *int   `json:"stock,omitempty"`
}

var products = []Product{{"mug", "Incident Response Mug", 1800, "USD", "☕", "", nil}, {"notebook", "Runbook Notebook", 1200, "USD", "📓", "", nil}, {"shirt", "Ship It T-Shirt", 2900, "USD", "👕", "", nil}, {"keyboard", "Deploy Keyboard", 7900, "USD", "⌨️", "", nil}}

func reply(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// Optional commerce source preserves the previous API during a staged rollout.
// Inventory and orders share one transactional owner; this service exposes the read catalogue.
func catalogueProducts() ([]Product, error) {
	base := os.Getenv("COMMERCE_URL")
	if base == "" {
		return products, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", base+"/catalogue/products", nil)
	if err != nil {
		return nil, err
	}
	response, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("catalogue unavailable")
	}
	var result []Product
	if err := json.NewDecoder(io.LimitReader(response.Body, 2*1024*1024)).Decode(&result); err != nil {
		return nil, err
	}
	return result, nil
}
func handler() http.Handler {
	m := http.NewServeMux()
	for _, p := range []string{"/health/live", "/health/ready"} {
		m.HandleFunc("GET "+p, func(w http.ResponseWriter, r *http.Request) { reply(w, 200, map[string]string{"status": "ok"}) })
	}
	m.HandleFunc("GET /products", func(w http.ResponseWriter, r *http.Request) {
		values, err := catalogueProducts()
		if err != nil {
			reply(w, 503, map[string]string{"error": "catalogue_unavailable"})
			return
		}
		reply(w, 200, values)
	})
	m.HandleFunc("GET /products/{id}", func(w http.ResponseWriter, r *http.Request) {
		values, err := catalogueProducts()
		if err != nil {
			reply(w, 503, map[string]string{"error": "catalogue_unavailable"})
			return
		}
		for _, p := range values {
			if p.ID == r.PathValue("id") {
				reply(w, 200, p)
				return
			}
		}
		reply(w, 404, map[string]string{"error": "product_not_found"})
	})
	m.HandleFunc("GET /metrics", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4")
		_, _ = w.Write([]byte(exposition()))
	})
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: 200}
		m.ServeHTTP(sw, r)
		observe(r, sw.status, time.Since(t).Seconds())
		slog.Info("request", "method", r.Method, "path", r.URL.Path, "duration_ms", time.Since(t).Milliseconds(), "request_id", r.Header.Get("X-Request-ID"))
	})
}
func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, nil)))
	s := http.Server{Addr: ":8081", Handler: handler(), ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	if err := s.ListenAndServe(); err != nil {
		slog.Error("server", "error", err)
		os.Exit(1)
	}
}
