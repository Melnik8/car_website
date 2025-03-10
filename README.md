# Cars-Viewer

Cars-Viewer is a simple web app that displays information about different car models using a Cars API. It helps users find and compare car specifications and manufacturers easily.

## Features

- **Display Car Info**: View details about various car models and their specifications.
- **Search and Filter**: Quickly search and filter cars by name, manufacturer, or category.
- **Compare Cars**: Compare different car models side-by-side.
- **Responsive Design**: Optimized for both desktop and mobile use.

## Technologies Used

- **Frontend**: HTML, CSS
- **Backend**: Go (Golang)
- **API**: Node.js
- **Data**: JSON format

## Setup and Installation

1. **Clone the repository**:
   ```bash
   git clone https://gitea.koodsisu.fi/juliageorgieva/cars
   cd cars
   ```
2. **Start the Cars API Server**:

    Ensure Node.js is installed. Download from Node.js.
    In the API directory, run:

    ```bash
    node main.js
    ```

3. **Start the Go Backend**:

    Ensure Go is installed. Download from Go.
    Run the backend server:
    ```bash
    go run main.go
    ```
3. **Open the App**:

    Visit http://localhost:8080 in your web browser.

## API Details
The Cars API provides car data in JSON format. 
    
## How to Use
- **Home Page**: Browse car models.
- **Search**: Use the search bar for specific car manufacturer, category, year and country (only).
- **Filter**: Apply filters by manufacturer, category, country or year.
- **Details**: Click on a car for more details.
    
## Project Structure
- **Backend**: main.go, main_test.go, structs.go
- **Frontend**: HTML files (index.html, etc.)
- **Styles**: styles.css
- **API Server**: main.js (Node.js)

## Contributing and License
Contributions are welcome! Please fork the repo and submit a pull request with your changes.
Free to use and explore. Don't forget to credit us if used. :)


