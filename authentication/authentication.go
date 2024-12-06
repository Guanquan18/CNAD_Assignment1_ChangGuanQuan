package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"io"
	"database/sql"
	"encoding/json"
	"github.com/gorilla/handlers"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
)

type User struct {
	UserId 			int
    Email 			string
	PasswordHash 	string
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/authenticate/login", loginUser).Methods("POST")

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

	fmt.Println("Listening at port 5000")
	log.Fatal(http.ListenAndServe(":5000", recoveryHandler))
}

func loginUser(w http.ResponseWriter, r *http.Request) {

	// Read request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}


	// Unmarshal the JSON byte array into a LoginRequest struct
	var loginRequest LoginRequest
	err = json.Unmarshal(jsonByte, &loginRequest)
	if err != nil {
		fmt.Println("Error unmarshalling JSON")
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}


	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/authentication_db")
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}

	// Check if the user exists in the database
	var query string = "SELECT * FROM Authentication WHERE Email = ? AND PasswordHash = ?"
	results, err := db.Query(query, loginRequest.Email, loginRequest.Password)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}

	// Create a slice to store the data and check if the user exists
	var user User
	if results.Next() {
		err = results.Scan(&user.UserId, &user.Email, &user.PasswordHash)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
			return
		}
	} else {
		fmt.Println("User not found")
		w.WriteHeader(http.StatusNotFound)	// Set the status code to 404
		return
	}

	defer db.Close()
	fmt.Println("User logged in successfully: ", user)
	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
	w.WriteHeader(http.StatusOK)	// Set the status code to 200
	json.NewEncoder(w).Encode(user)	// Encode the user into JSON and write it to the response
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}

	// Execute the query
	var query string = "SELECT * FROM User"
	results, err := db.Query(query)
	if err != nil {
		fmt.Println(err.Error())
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		return
	}

	// Create a slice to store the data
	var UserMap map[int]User = make(map[int]User)
	for results.Next() {

		var user User
		err = results.Scan(&user.UserId, &user.Email, &user.PasswordHash)
		if err != nil {
			fmt.Println(err.Error())
			w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
			return
		}
		UserMap[user.UserId] = user	// Add the user to the map
	}

	defer db.Close()

	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
	w.WriteHeader(http.StatusOK)	// Set the status code to 200
	json.NewEncoder(w).Encode(UserMap)	// Encode the map into JSON and write it to the response
}