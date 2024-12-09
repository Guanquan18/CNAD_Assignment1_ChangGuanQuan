package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"io"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Define a struct to hold reservation data
type Reservation struct {
	ReservationId int    	`json:"reservationId"`
	UserId        int    	`json:"userId"`
	VehicleId     int    	`json:"vehicleId"`
	StartDate     string 	`json:"startDate"`
	EndDate       string 	`json:"endDate"`
	TotalCost     float64   `json:"totalCost"`
	Status        string 	`json:"status"`
}

type ErrorResponse struct {
	Message string `json:"Message"`
}

func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/reservation", searchReservation).Methods("GET")
	router.HandleFunc("/reservation/make-reservation", createReservation).Methods("POST")
	router.HandleFunc("/reservation/upcoming/{userId}", getReservation).Methods("GET")
	router.HandleFunc("/reservation/check-reservation-limit/{userId}", checkReservationLimits).Methods("POST")
	router.HandleFunc("/reservation/history/{userId}", getReservation).Methods("GET")
	router.HandleFunc("/reservation/cancel/{reservationId}", cancelReservation).Methods("PUT")
	router.HandleFunc("/reservation/update/{reservationId}", updateReservation).Methods("PUT")

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

	fmt.Println("Listening at port 5003")
	log.Fatal(http.ListenAndServe(":5003", recoveryHandler))
}

func checkReservationLimits(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	defer db.Close()

	// Read the body of the request
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Unmarshal the request body into a struct
	type checkReservationRequest struct {
		BookingLimit int `json:"BookingLimit"`
	}

	var req checkReservationRequest
	err = json.Unmarshal(jsonByte, &req)
	if err != nil {
		fmt.Println("Error unmarshalling request body:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Get the user ID from the URL parameters
	params := mux.Vars(r)
	userId := params["userId"]

	// Query the database to get the number of upcoming reservations for the user
	rows, err := db.Query("SELECT COUNT(*) FROM Reservation WHERE UserId = ? AND StartDate >= CURDATE() AND Status = 'Confirmed'", userId)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer rows.Close()

	var count int
	if rows.Next() {
		err := rows.Scan(&count)
		if err != nil {
			fmt.Println("Error scanning reservation count:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}
	}

	// Check if the user has exceeded the booking limit
	if count >= req.BookingLimit {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable)
	}else{
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

func updateReservation(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	defer db.Close()

	// Get the reservation ID from the URL parameters
	params := mux.Vars(r)
	reservationId := params["reservationId"]

	// Read the body of the request
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Unmarshal the request body into a struct
	type updateReservationRequest struct {
		StartDate string `json:"StartDate"`
		EndDate   string `json:"EndDate"`
		VehicleId int    `json:"VehicleId"`
	}

	var req updateReservationRequest
	err = json.Unmarshal(jsonByte, &req)
	if err != nil {
		fmt.Println("Error unmarshalling request body:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Check if the another reservation exists for the same vehicle and time period
	rows, err := db.Query("SELECT * FROM Reservation WHERE VehicleId = ? AND StartDate < ? AND EndDate > ? AND Status = 'Confirmed' AND ReservationId != ?", req.VehicleId, req.EndDate, req.StartDate, reservationId)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// If another reservation exists, return a conflict response
	if rows.Next() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusConflict)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Another reservation exists for the same vehicle and time period"})
		return
	}

	// Update the reservation in the database
	_, err = db.Exec("UPDATE Reservation SET StartDate = ?, EndDate = ? WHERE ReservationId = ?", req.StartDate, req.EndDate, reservationId)
	if err != nil {
		fmt.Println("Error updating reservation:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return the reservation details as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
}


func cancelReservation(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	defer db.Close()

	// Get the reservation ID from the URL parameters
	params := mux.Vars(r)
	reservationId := params["reservationId"]

	// Update the reservation status to "Cancelled"
	_, err = db.Exec("UPDATE Reservation SET Status = 'Cancelled' WHERE ReservationId = ?", reservationId)
	if err != nil {
		fmt.Println("Error updating reservation status:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Reservation cancelled successfully"})
}


func getReservation(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	defer db.Close()

	// Get the user ID from the URL parameters
	params := mux.Vars(r)
	userId := params["userId"]

	// check if url query parameter is upcoming or history
	urlQuery := r.URL.Query()
	var query string
	if urlQuery.Get("type") == "history" {
		// Query the database to get past reservations for the user
		query = "SELECT * FROM Reservation WHERE UserId = ? AND EndDate < CURDATE() AND Status = 'Confirmed'"
	}else if urlQuery.Get("type") == "upcoming" {
		// Query the database to get upcoming reservations for the user
		query = "SELECT * FROM Reservation WHERE UserId = ? AND StartDate >= CURDATE() AND Status = 'Confirmed'"
	}


	// Query the database to get upcoming reservations for the user
	rows, err := db.Query(query, userId)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer rows.Close()

	var reservations []Reservation
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ReservationId, &reservation.UserId, &reservation.VehicleId, &reservation.StartDate, &reservation.EndDate, &reservation.TotalCost, &reservation.Status)
		if err != nil {
			fmt.Println("Error scanning reservation row:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}
		reservations = append(reservations, reservation)
	}

	if len(reservations) == 0 {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]string{}) // Return an empty array if no reservations
		return
	}

	// Define a struct to hold vehicle details
	type VehicleDetails struct {
		Make     string `json:"make"`
		Model    string `json:"model"`
		Location string `json:"location"`
	}

	// Fetch vehicle details for each reservation
	vehicleDetailsMap := make(map[int]VehicleDetails)
	for _, reservation := range reservations {
		vehicleURL := fmt.Sprintf("http://localhost:5002/vehicle/get-vehicle/%d", reservation.VehicleId)
		resp, err := http.Get(vehicleURL)
		if err != nil {
			fmt.Println("Error fetching vehicle details:", err)
			continue // Skip this vehicle and log the error
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			fmt.Printf("Vehicle service returned status %d for vehicleId %d\n", resp.StatusCode, reservation.VehicleId)
			continue
		}

		var vehicleDetails VehicleDetails
		err = json.NewDecoder(resp.Body).Decode(&vehicleDetails)
		if err != nil {
			fmt.Println("Error decoding vehicle details:", err)
			continue
		}
		vehicleDetailsMap[reservation.VehicleId] = vehicleDetails
	}

	// Combine reservations and vehicle details
	type CombinedResponse struct {
		Reservation Reservation   `json:"reservation"`
		Vehicle     VehicleDetails `json:"vehicle"`
	}

	var combinedReservations []CombinedResponse
	for _, reservation := range reservations {
		if vehicle, found := vehicleDetailsMap[reservation.VehicleId]; found {
			combinedReservations = append(combinedReservations, CombinedResponse{
				Reservation: reservation,
				Vehicle:     vehicle,
			})
		} else {
			fmt.Printf("No vehicle details found for vehicleId %d\n", reservation.VehicleId)
		}
	}

	// Send the combined data to the frontend
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(combinedReservations)
}


func createReservation(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}

	// Decode the request body into a struct
	type reservationRequest struct {
		UserId      int    		`json:"UserId"`
		VehicleId   int    		`json:"VehicleId"`
		StartDate   string 		`json:"StartDate"`
		EndDate     string 		`json:"EndDate"`
		TotalCost   float64 	`json:"TotalCost"`
	}

	// Read request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Unmarshal the request body into a struct
	var req reservationRequest
	err = json.Unmarshal(jsonByte, &req)
	if err != nil {
		fmt.Println("Error unmarshalling request body:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}

	// Insert the reservation into the database
	result, err := db.Exec("INSERT INTO Reservation (VehicleId, UserId, StartDate, EndDate, TotalCost) VALUES (?, ?, ?, ?, ?)", req.VehicleId, req.UserId, req.StartDate, req.EndDate, req.TotalCost)
	if err != nil {
		fmt.Println("Error inserting reservation into the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Get the auto-generated reservation ID
	reservationId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting the reservation ID:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return the reservation ID as JSON
	type Response struct {
		ReservationId int `json:"ReservationId"`
	}

	var response Response
	response.ReservationId = int(reservationId)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}


func searchReservation(w http.ResponseWriter, r *http.Request) {
	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/reservation_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	defer db.Close()

	// Get starting and ending date from the query parameters
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")

	if startDate == "" || endDate == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Missing date range parameters"})
		return
	}

	// Query the database to get reservations within the specified date range
	rows, err := db.Query("SELECT * FROM Reservation WHERE StartDate < ? AND EndDate > ? AND Status='Confirmed'", endDate, startDate)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer rows.Close()

	var reservations []Reservation
	for rows.Next() {
		var reservation Reservation
		err := rows.Scan(&reservation.ReservationId, &reservation.VehicleId, &reservation.UserId, &reservation.StartDate, &reservation.EndDate, &reservation.TotalCost, &reservation.Status)
		if err != nil {
			fmt.Println("Error scanning reservation row:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}
		reservations = append(reservations, reservation)
	}

	// Return reservations as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(reservations)
}
