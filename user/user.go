package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"io"
	"strconv"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/handlers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	UserId 			int
    Email 			string
    FirstName 		string
    LastName 		string
    MembershipTier 	string
}

type ErrorResponse struct {
	Message string `json:"Message"`
}

type UpdateUserRequest struct {
	Email 			string `json:"Email"`
	FirstName 		string `json:"FirstName"`
	LastName 		string `json:"LastName"`
}

func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/account/{userId}", GetUser).Methods("GET")
	router.HandleFunc("/account/update/{userId}", UpdateUser).Methods("PUT")

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

	fmt.Println("Listening at port 5001")
	log.Fatal(http.ListenAndServe(":5001", recoveryHandler))
}

func UpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	// convert the userId to an integer
    userId, err := strconv.Atoi(vars["userId"])
    if err != nil {
        fmt.Println("Error converting userId to integer")
        w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
        return
    }

	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		fmt.Println("Error connecting to the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Read request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Unmarshal the JSON byte array into a User struct
	var user UpdateUserRequest
	err = json.Unmarshal(jsonByte, &user)
	if err != nil {
		fmt.Println("Error unmarshalling JSON")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Check if the user exists in the database
	var checkQuery string = "SELECT * FROM User WHERE UserId = ?"
	row := db.QueryRow(checkQuery, userId)
	var existingUser User
	err = row.Scan(&existingUser.UserId, &existingUser.Email, &existingUser.FirstName, &existingUser.LastName, &existingUser.MembershipTier)
	if err != nil {
		fmt.Println("Error checking if user exists in the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Check if the user exists
	if existingUser.UserId == 0 {
		fmt.Println("User already exists in the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotAcceptable) // Set the status code to 406
		json.NewEncoder(w).Encode(ErrorResponse{Message: "User does not exist"})
		return
	} 

	// Update the user in the database
	var updateQuery string = "UPDATE User SET Email = ?, FirstName = ?, LastName = ? WHERE UserId = ?"
	_, err = db.Exec(updateQuery, user.Email, user.FirstName, user.LastName, userId)
	if err != nil {
		fmt.Println("Error updating user in the database")
		fmt.Println(err.Error())

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Retrieve the updated user from the database
	var selectQuery string = "SELECT * FROM User WHERE UserId = ?"
	row = db.QueryRow(selectQuery, userId)
	var updatedUser User
	err = row.Scan(&updatedUser.UserId, &updatedUser.Email, &updatedUser.FirstName, &updatedUser.LastName, &updatedUser.MembershipTier)
	if err != nil {
		fmt.Println("Error retrieving updated user from the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	db.Close()

	// Return the updated user as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // Set the status code to 202
	json.NewEncoder(w).Encode(updatedUser)

}

func GetUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	userId := vars["userId"]
	

	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		fmt.Println("Error connecting to the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Query the database for the user with the given ID
	var query string = "SELECT * from User where UserId = ?;"
	results, err := db.Query(query, userId)
	if err != nil {
		fmt.Println("Error querying the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Create a User struct to hold the user data
	var user User

	// Iterate over the query results
	for results.Next() {
		err = results.Scan(&user.UserId, &user.Email, &user.FirstName, &user.LastName, &user.MembershipTier)
		if err != nil {
			fmt.Println("Error scanning query results")
			w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
			return
		}
	}

	db.Close()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Set the status code to 200
	json.NewEncoder(w).Encode(user)
}