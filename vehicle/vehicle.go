package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"encoding/json"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Define a struct to hold vehicle data
type Vehicle struct {
	VehicleId   int     `json:"vehicleId"`
	VehicleType string  `json:"vehicleType"`
	Make        string  `json:"make"`
	Model       string  `json:"model"`
	Year        int     `json:"year"`
	Location    string  `json:"location"`
	HourlyRate  float64 `json:"hourlyRate"`
	Availability bool   `json:"availability"`
}

type ErrorResponse struct {
	Message string `json:"Message"`
}

func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/vehicle", searchVehicle).Methods("GET")
	router.HandleFunc("/vehicle/get-vehicle/{vehicleId}", getVehicle).Methods("GET")

	// Set up logging middleware to log requests to the console
	loggingHandler := handlers.LoggingHandler(os.Stdout, router)

	// Set up CORS middleware to allow all origins to access the API
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}), // Replace "*" with specific domains as needed
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(loggingHandler)

	// Recovery handler to recover from panics and log them
	recoveryHandler := handlers.RecoveryHandler(handlers.PrintRecoveryStack(true))(corsHandler)

	fmt.Println("Listening at port 5002")
	log.Fatal(http.ListenAndServe(":5002", recoveryHandler))
}

func getVehicle(w http.ResponseWriter, r *http.Request) {
	// Connect to the Vehicle database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/vehicle_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	defer db.Close()

	// Get the vehicle ID from the URL path
	vars := mux.Vars(r)
	vehicleId := vars["vehicleId"]

	// Query the database to get the vehicle with the specified ID
	row := db.QueryRow("SELECT Make, Model, Location FROM Vehicle WHERE VehicleId = ?", vehicleId)

	type VehicleDetails struct {
		Make     string `json:"Make"`
		Model    string `json:"Model"`
		Location string `json:"Location"`
	}

	var vehicleDetails VehicleDetails
	err = row.Scan(&vehicleDetails.Make, &vehicleDetails.Model, &vehicleDetails.Location)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return the vehicle details as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(vehicleDetails)
}



func searchVehicle(w http.ResponseWriter, r *http.Request) {
	// Connect to the Vehicle database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/vehicle_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer db.Close()

	// Get query parameters for startDate and endDate
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	// Validate query parameters
	if startDate == "" || endDate == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Missing time range parameters"})
		return
	}

	// Fetch all vehicles from the Vehicles table
	rows, err := db.Query("SELECT * FROM Vehicle")
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer rows.Close()

	var vehicles []Vehicle // Store all vehicles

	// Iterate through vehicles and populate the slice
	for rows.Next() {
		var vehicle Vehicle
		err := rows.Scan(&vehicle.VehicleId, &vehicle.VehicleType, &vehicle.Make, &vehicle.Model, &vehicle.Year, &vehicle.Location, &vehicle.HourlyRate, &vehicle.Availability)
		if err != nil {
			fmt.Println("Error scanning row:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}
		vehicles = append(vehicles, vehicle)
	}


	// Fetch unavailable vehicles from the Reservation server
	reservationURL := fmt.Sprintf("http://localhost:5003/reservation?startDate=%s&endDate=%s", startDate, endDate)
	resp, err := http.Get(reservationURL)
	if err != nil {
		fmt.Println("Error calling Reservation server:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error response from Reservation server:", resp.Status)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error fetching reservations"})
		return
	}

	type Reservation struct {
		ReservationId int `json:"reservationId"`
		VehicleId     int `json:"vehicleId"`
		UserId        int `json:"userId"`
		StartDate     string `json:"startDate"`
		EndDate       string `json:"endDate"`
		TotalAmount   float64 `json:"totalAmount"`
	}

	// Parse response from Reservation server
	var reservations []Reservation
	err = json.NewDecoder(resp.Body).Decode(&reservations)
	if err != nil {
		fmt.Println("Error decoding response from Reservation server:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Filter vehicles based on reservations
	unavailableVehicleIds := make(map[int]bool)
	for _, reservation := range reservations {
		unavailableVehicleIds[reservation.VehicleId] = true
	}

	availableVehicles := []Vehicle{}
	for _, vehicle := range vehicles {
		if !unavailableVehicleIds[vehicle.VehicleId] {
			availableVehicles = append(availableVehicles, vehicle)
		}
	}

	// Return the available vehicles as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(availableVehicles)
}



