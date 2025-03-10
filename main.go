package main

import (
	"cars/structs"
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type App struct {
	templates     *template.Template
	manufacturers []structs.Manufacturer
	carModels     []structs.CarModel
	categories    []structs.Category
}

func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

func main() {
	funcMap := template.FuncMap{
		"contains": contains,
	}

	app := &App{
		templates: template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html")),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", app.indexHandler)
	mux.HandleFunc("/error", app.errorHandler)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	mux.Handle("/img/", http.StripPrefix("/img/", http.FileServer(http.Dir("api/img"))))
	mux.HandleFunc("/car", app.CarDetailsHandler)
	mux.HandleFunc("/favicon.png", app.faviconHandler)
	mux.HandleFunc("/notfound", app.notFoundHandler)
	mux.HandleFunc("/health", app.healthCheckHandler)
	mux.HandleFunc("/filter", app.filterHandler)
	mux.HandleFunc("/search", app.searchHandler)
	mux.HandleFunc("/compare", app.compareHandler)

	app.loadData()

	wrappedMux := app.errorHandlerMiddleware(app.catchAllHandler(mux))

	log.Println("Server is running on http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", wrappedMux))
}

func (app *App) indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if r.URL.Path != "/" {
		app.notFoundHandler(w, r)
		return
	}

	if err := app.loadData(); err != nil {
		log.Printf("Failed to load data: %v", err)
		app.renderError(w, http.StatusInternalServerError, "Could not connect to the API server. Please try again later.")
		return
	}

	manufacturersMap := make(map[string]string)
	for _, manufacturer := range app.manufacturers {
		manufacturersMap[strconv.Itoa(manufacturer.ID)] = manufacturer.Name
	}

	data := structs.PageData{
		Title:                 "Aurora cars",
		Manufacturers:         app.manufacturers,
		CarModels:             app.carModels,
		Categories:            app.categories,
		Countries:             app.getUniqueCountries(app.manufacturers),
		Years:                 app.getUniqueYears(app.carModels),
		SelectedManufacturers: []string{},
		SelectedCategories:    []string{},
		SelectedYears:         []string{},
		SelectedCountries:     []string{},
		ManufacturersMap:      manufacturersMap,
	}

	if err := app.templates.ExecuteTemplate(w, "layout.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}
func (app *App) checkAPIHealth() {
	ticker := time.NewTicker(5 * time.Minute)
	for range ticker.C {
		_, err := http.Get("http://localhost:3000/health")
		if err != nil {
			log.Println("API server is down:", err)
		} else {
			log.Println("API server is up and running")
		}
	}
}

func (app *App) CarDetailsHandler(w http.ResponseWriter, r *http.Request) {
	carIDStr := r.URL.Query().Get("id")
	carID, err := strconv.Atoi(carIDStr)

	if err != nil {
		log.Printf("Error parsing car ID '%s': %v", carIDStr, err)
		http.Error(w, "Invalid car ID format: %s", http.StatusInternalServerError)
		return
	}

	var car *structs.CarModel
	var manData *structs.Manufacturer
	for _, c := range app.carModels {
		if c.ID == carID {
			car = &c
			for _, m := range app.manufacturers {
				if car.ManufacturerID == m.ID {
					manData = &m
					break
				}
			}
			break
		}
	}

	if car == nil || manData == nil {
		http.Error(w, "Car or Manufacturer not found", http.StatusNotFound)
		return
	}

	data := struct {
		Car     *structs.CarModel
		ManData *structs.Manufacturer
	}{
		Car:     car,
		ManData: manData,
	}

	if err := app.templates.ExecuteTemplate(w, "car.html", data); err != nil {
		log.Printf("Error executing template: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *App) errorHandlerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("Panic recovered: %v", err)
				app.errorHandler(w, r)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func (app *App) recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				log.Printf("Recovered from panic: %v", err)
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func fetchDataWithTimeout(url string) ([]byte, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Get(url)
	if err != nil {
		log.Println("Failed to fetch data:", err)
		return nil, err
	}
	defer resp.Body.Close()

	return io.ReadAll(resp.Body)
}

func (app *App) errorHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "500 - Internal Server Error. We're sorry, but something went wrong. Please try again later.", http.StatusInternalServerError)
}

func (app *App) notFoundHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "404 - Page Not Found. The page you are looking for does not exist.", http.StatusNotFound)
}

func (app *App) catchAllHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path != "/" && path != "/favicon.png" && !strings.HasPrefix(path, "/static/") && !strings.HasPrefix(path, "/img/") && path != "/error" && path != "/notfound" && path != "/car" && path != "/filter" && path != "/search"  && path != "/compare" {
			app.notFoundHandler(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (app *App) faviconHandler(w http.ResponseWriter, r *http.Request) {
	if _, err := os.Stat("static/favicon.png"); os.IsNotExist(err) {
		log.Printf("favicon.png not found: %v", err)
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, "static/favicon.png")
}

func (app *App) healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	select {
	case <-r.Context().Done():
		log.Println("Health check request cancelled")
		return
	default:
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}

func (app *App) loadDataPeriodically() {
	ticker := time.NewTicker(30 * time.Minute)
	go func() {
		for {
			<-ticker.C
			app.loadData()
		}
	}()
	app.loadData()
}

func (app *App) loadData() error {
	client := &http.Client{Timeout: 10 * time.Second}
	errorsChan := make(chan error, 3)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		err := app.fetchData("http://localhost:3000/api/manufacturers", &app.manufacturers, client)
		errorsChan <- err
	}()

	go func() {
		defer wg.Done()
		err := app.fetchData("http://localhost:3000/api/models", &app.carModels, client)
		errorsChan <- err
	}()

	go func() {
		defer wg.Done()
		err := app.fetchData("http://localhost:3000/api/categories", &app.categories, client)
		errorsChan <- err
	}()

	go func() {
		wg.Wait()
		close(errorsChan)
	}()

	for err := range errorsChan {
		if err != nil {
			return err
		}
	}

	log.Println("Data loaded successfully from all APIs")
	return nil
}

func (app *App) fetchData(url string, target interface{}, client *http.Client) error {
	resp, err := client.Get(url)
	if err != nil {
		//log.Printf("Failed to fetch data from %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		log.Printf("API returned non-200 status from %s: %d, Response: %s", url, resp.StatusCode, string(bodyBytes))
		return fmt.Errorf("API returned non-200 status from %s: %d", url, resp.StatusCode)
	}

	err = json.NewDecoder(resp.Body).Decode(target)
	if err != nil {
		log.Printf("Failed to decode data from %s: %v", url, err)
		return err
	}
	return nil
}

func fetchData(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Failed to fetch data:", err)
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
func (app *App) renderError(w http.ResponseWriter, statusCode int, message string) {
	w.WriteHeader(statusCode)
	data := struct {
		Title        string
		ErrorMessage string
	}{
		Title:        "Error - Aurora Cars",
		ErrorMessage: message,
	}
	err := app.templates.ExecuteTemplate(w, "layout.html", data)
	if err != nil {
		log.Printf("Error executing template for error: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

func (app *App) loadManufacturers(wg *sync.WaitGroup, client *http.Client, ch chan error) {
	defer wg.Done()
	start := time.Now()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:3000/api/manufacturers", nil)
	if err != nil {
		ch <- err
		return
	}

	resp, err := client.Do(req)
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- err
		return
	}

	if err := json.Unmarshal(body, &app.manufacturers); err != nil {
		ch <- err
		return
	}

	ch <- nil
	duration := time.Since(start)
	log.Printf("Loading manufacturers took %v", duration)
}

func (app *App) loadCarModels(wg *sync.WaitGroup, client *http.Client, ch chan error) {
	defer wg.Done()
	start := time.Now()
	resp, err := client.Get("http://localhost:3000/api/models")
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- err
		return
	}

	if err := json.Unmarshal(body, &app.carModels); err != nil {
		ch <- err
		return
	}

	ch <- nil
	duration := time.Since(start)
	log.Printf("Loading car models took %v", duration)
}

func (app *App) loadCategories(wg *sync.WaitGroup, client *http.Client, ch chan error) {
	defer wg.Done()
	start := time.Now()
	resp, err := client.Get("http://localhost:3000/api/categories")
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- err
		return
	}

	if err := json.Unmarshal(body, &app.categories); err != nil {
		ch <- err
		return
	}

	ch <- nil
	duration := time.Since(start)
	log.Printf("Loading categories took %v", duration)
}

func (app *App) getUniqueCountries(manufacturers []structs.Manufacturer) []string {
	uniqueCountries := make(map[string]bool)
	var countries []string

	for _, manufacturer := range manufacturers {
		if !uniqueCountries[manufacturer.Country] {
			uniqueCountries[manufacturer.Country] = true
			countries = append(countries, manufacturer.Country)
		}
	}

	return countries
}

func (app *App) getUniqueYears(carModels []structs.CarModel) []int {
	uniqueYears := make(map[int]bool)
	var years []int

	for _, car := range carModels {
		if !uniqueYears[car.Year] {
			uniqueYears[car.Year] = true
			years = append(years, car.Year)
		}
	}

	sort.Ints(years)
	return years
}

func (app *App) filterHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	manufacturerID := r.FormValue("manufacturer")
	categoryID := r.FormValue("category")
	year := r.FormValue("year")
	country := r.FormValue("country")

	filteredCars := []structs.CarModel{}
	for _, car := range app.carModels {
		if manufacturerID != "" && strconv.Itoa(car.ManufacturerID) != manufacturerID {
			continue
		}
		if categoryID != "" && strconv.Itoa(car.CategoryID) != categoryID {
			continue
		}
		if year != "" && strconv.Itoa(car.Year) != year {
			continue
		}
		if country != "" && !app.isCarFromCountry(car, country) {
			continue
		}
		filteredCars = append(filteredCars, car)
	}

	data := structs.PageData{
		Title:                 "Aurora cars",
		Manufacturers:         app.manufacturers,
		CarModels:             filteredCars,
		Categories:            app.categories,
		Countries:             app.getUniqueCountries(app.manufacturers),
		Years:                 app.getUniqueYears(app.carModels),
		SelectedManufacturers: []string{manufacturerID},
		SelectedCategories:    []string{categoryID},
		SelectedYears:         []string{year},
		SelectedCountries:     []string{country},
		NoResults:             len(filteredCars) == 0,
	}

	app.templates.ExecuteTemplate(w, "layout.html", data)
}

func (app *App) isCarFromCountry(car structs.CarModel, country string) bool {
	for _, manufacturer := range app.manufacturers {
		if manufacturer.ID == car.ManufacturerID && manufacturer.Country == country {
			return true
		}
	}
	return false
}

func secureHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'")
		w.Header().Set("X-Frame-Options", "DENY")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "no-referrer")

		next.ServeHTTP(w, r)
	})
}

func (app *App) searchHandler(w http.ResponseWriter, r *http.Request) {
	query := strings.ToLower(r.URL.Query().Get("query"))
	var results []structs.CarModel

	for _, car := range app.carModels {
		manName := app.getManufacturerNameByID(car.ManufacturerID)
		catName := app.getCategoryNameByID(car.CategoryID)

		country := app.getCountryByManufacturerID(car.ManufacturerID)

		searchText := strings.ToLower(
			car.Name + " " +
				strconv.Itoa(car.Year) + " " +
				manName + " " +
				catName + " " +
				country,
		)

		if strings.Contains(searchText, query) {
			results = append(results, car)
		}
	}

	data := structs.PageData{
		Title:         "Search Results",
		CarModels:     results,
		Manufacturers: app.manufacturers,
		Categories:    app.categories,
		Countries:     app.getUniqueCountries(app.manufacturers),
		Years:         app.getUniqueYears(app.carModels),
		Query:         query,
	}
	app.templates.ExecuteTemplate(w, "layout.html", data)
}

func (app *App) getCountryByManufacturerID(id int) string {
	for _, m := range app.manufacturers {
		if m.ID == id {
			return m.Country
		}
	}
	return ""
}

func (app *App) getManufacturerNameByID(id int) string {
	for _, m := range app.manufacturers {
		if m.ID == id {
			return m.Name
		}
	}
	return ""
}

func (app *App) getCategoryNameByID(id int) string {
	for _, cat := range app.categories {
		if cat.ID == id {
			return cat.Name
		}
	}
	return ""
}

func (app *App) getManufacturerCountry(manufacturerID int) string {
	for _, manufacturer := range app.manufacturers {
		if manufacturer.ID == manufacturerID {
			return manufacturer.Country
		}
	}
	return ""
}

func (app *App) compareHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	carIDs := r.Form["car_ids"] 

	var carsToCompare []structs.CarModel
	for _, idStr := range carIDs {
		id, err := strconv.Atoi(idStr)
		if err == nil {
			for _, car := range app.carModels {
				if car.ID == id {
					carsToCompare = append(carsToCompare, car)
				}
			}
		}
	}

	manuMap := make(map[int]structs.Manufacturer)
	for _, car := range carsToCompare {
		for _, manufacturer := range app.manufacturers {
			if manufacturer.ID == car.ManufacturerID {
				manuMap[car.ManufacturerID] = manufacturer
				break
			}
		}
	}

	data := structs.PageData{
		Title:     "Car Comparison",
		CarModels: carsToCompare,
		ManuMap:   manuMap,
	}

	app.templates.ExecuteTemplate(w, "compare.html", data)
}