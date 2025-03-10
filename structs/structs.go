package structs

type Manufacturer struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Country string `json:"country"`
	Founded int    `json:"foundingYear"`
}

type Specifications struct {
	Engine       string `json:"engine"`
	Horsepower   int    `json:"horsepower"`
	Transmission string `json:"transmission"`
	Drivetrain   string `json:"drivetrain"`
}

type CarModel struct {
	ID               int            `json:"id"`
	Name             string         `json:"name"`
	ManufacturerID   int            `json:"manufacturerId"`
	CategoryID       int            `json:"categoryId"`
	Year             int            `json:"year"`
	Specifications   Specifications `json:"specifications"`
	Image            string         `json:"image"`
	ManufacturerName string         `json:"manufacturerName,omitempty"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type PageData struct {
	Title                 string
	Manufacturers         []Manufacturer
	CarModels             []CarModel
	FilteredCars          []CarModel
	Categories            []Category
	Countries             []string
	Years                 []int
	SelectedManufacturers []string
	SelectedCategories    []string
	SelectedYears         []string
	SelectedCountries     []string
	ManufacturersMap      map[string]string
	Results               []CarModel
	Query                 string
	NoResults             bool
	ErrorMessage          string
	ManuMap               map[int]Manufacturer
}
