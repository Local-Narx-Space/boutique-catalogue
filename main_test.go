package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestProducts(t *testing.T) {
	w := httptest.NewRecorder()
	handler().ServeHTTP(w, httptest.NewRequest("GET", "/products", nil))
	var p []Product
	if json.Unmarshal(w.Body.Bytes(), &p) != nil || len(p) != 4 || w.Code != 200 {
		t.Fatal(w.Body.String())
	}
	for _, v := range p {
		if v.Price <= 0 {
			t.Fatal("invalid price")
		}
	}
}
func TestMissing(t *testing.T) {
	w := httptest.NewRecorder()
	handler().ServeHTTP(w, httptest.NewRequest("GET", "/products/missing", nil))
	if w.Code != 404 {
		t.Fatal(w.Code)
	}
}

func TestCommerceCatalogue(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/catalogue/products" {
			t.Error(r.URL.Path)
		}
		_, _ = w.Write([]byte(`[{"id":"new-product","name":"New","price":109900,"currency":"INR","emoji":"◈","stock":3}]`))
	}))
	defer source.Close()
	t.Setenv("COMMERCE_URL", source.URL)
	response := httptest.NewRecorder()
	handler().ServeHTTP(response, httptest.NewRequest("GET", "/products/new-product", nil))
	var product Product
	if response.Code != 200 || json.Unmarshal(response.Body.Bytes(), &product) != nil || product.ID != "new-product" || product.Stock == nil || *product.Stock != 3 {
		t.Fatal(response.Body.String())
	}
}

func TestCommerceFailureDoesNotServeStalePrices(t *testing.T) {
	source := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(503) }))
	defer source.Close()
	t.Setenv("COMMERCE_URL", source.URL)
	response := httptest.NewRecorder()
	handler().ServeHTTP(response, httptest.NewRequest("GET", "/products", nil))
	if response.Code != 503 {
		t.Fatal(response.Code)
	}
}
