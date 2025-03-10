package main

import (
	"cars/structs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func setupTestRequest(method, url string) (*http.Request, *httptest.ResponseRecorder, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, nil, err
	}
	rr := httptest.NewRecorder()
	return req, rr, nil
}

func TestIndexHandler(t *testing.T) { 
	req, err := http.NewRequest("GET", "/", nil) 
		t.Fatal(err)
	}

	rr := httptest.NewRecorder() 
	handler := http.HandlerFunc(indexHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK { 
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK) 
	}

	if !strings.Contains(rr.Body.String(), "Car Viewer") { 
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String()) 
	}
}

func TestErrorHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/error", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(errorHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusInternalServerError)
	}

	if !strings.Contains(rr.Body.String(), "500 - Internal Server Error") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestNotFoundHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/notfound", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(notFoundHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusNotFound {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusNotFound)
	}

	if !strings.Contains(rr.Body.String(), "404 - Page Not Found") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestHealthCheckHandler(t *testing.T) {
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "OK") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}
func TestCarDetailsHandler_ValidID(t *testing.T) {
	app := &App{
		carModels: []structs.CarModel{
			{ID: 1, Name: "Test Car", ManufacturerID: 1},
		},
		manufacturers: []structs.Manufacturer{
			{ID: 1, Name: "Test Manufacturer"},
		},
	}

	req, rr, err := setupTestRequest("GET", "/car?id=1")
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(app.CarDetailsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "Test Car") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestCarDetailsHandler_InvalidID(t *testing.T) {
	app := &App{}

	req, rr, err := setupTestRequest("GET", "/car?id=invalid")
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(app.CarDetailsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
	}
}

func TestCarDetailsHandler_InvalidID_Format(t *testing.T) {
	app := setupApp()
	req, rr := httptest.NewRequest("GET", "/car?id=abc123", nil), httptest.NewRecorder()
	handler := http.HandlerFunc(app.CarDetailsHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("Expected HTTP 400 status, got: %v", status)
	}
	if !strings.Contains(rr.Body.String(), "Invalid car ID") {
		t.Errorf("Expected error message about invalid car ID, got: %v", rr.Body.String())
	}
}

func TestFilterHandler_NoResults(t *testing.T) {
	app := &App{
		carModels: []structs.CarModel{
			{ID: 1, Name: "Car A", ManufacturerID: 1, CategoryID: 1, Year: 2020},
		},
		manufacturers: []structs.Manufacturer{
			{ID: 1, Name: "Manufacturer A"},
		},
		categories: []structs.Category{
			{ID: 1, Name: "SUV"},
		},
	}

	req, rr, err := setupTestRequest("GET", "/filter?manufacturer=2")
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(app.filterHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "No results found") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestSearchHandler_NoResults(t *testing.T) {
	app := &App{
		carModels: []structs.CarModel{
			{ID: 1, Name: "Car A"},
		},
	}

	req, rr, err := setupTestRequest("GET", "/search?query=NotExist")
	if err != nil {
		t.Fatal(err)
	}

	handler := http.HandlerFunc(app.searchHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	if !strings.Contains(rr.Body.String(), "No results found") {
		t.Errorf("handler returned unexpected body: got %v", rr.Body.String())
	}
}

func TestInvalidPath(t *testing.T) {
	req, rr := httptest.NewRequest("GET", "/nonexistent", nil), httptest.NewRecorder()
	handler := http.HandlerFunc(app.notFoundHandler)
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("Expected status 404 for non-existent path, got %v", rr.Code)
	}
}
