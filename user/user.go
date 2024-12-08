package main

import (
	"fmt"
	"net/http"
	"os"
	"log"
	"io"
	"bytes"
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


func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/account/{userId}", getUser).Methods("GET")
	router.HandleFunc("/account/update/{userId}", updateUser).Methods("PUT")
	router.HandleFunc("/account/check-user/{userId}", checkUserExists).Methods("POST")
	router.HandleFunc("/account/check-user", checkUserExists).Methods("POST")
	router.HandleFunc("/account/create-user", createUser).Methods("POST")

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

func createUser(w http.ResponseWriter, r *http.Request) {

	type CreateUserRequest struct {
		Email 			string `json:"Email"`
		FirstName 		string `json:"FirstName"`
		LastName 		string `json:"LastName"`
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

	// Unmarshal the JSON byte array into a User struct
	var user CreateUserRequest
	err = json.Unmarshal(jsonByte, &user)
	if err != nil {
		fmt.Println("Error unmarshalling JSON", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request format"})
		return
	}

	// Connect to the database
	db, err := sql.Open("mysql",  "root:S10257825A@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		fmt.Println("Error connecting to the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Insert the user into the database
	var insertQuery string = "INSERT INTO User (Email, FirstName, LastName, MembershipTier) VALUES (?, ?, ?, 'Basic')"
	result, err := db.Exec(insertQuery, user.Email, user.FirstName, user.LastName)
	if err != nil {
		fmt.Println("Error inserting user into the database")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Retrieve the auto-generated ID of the new user
	userId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error retrieving user ID")

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Put the userId into a json object
	userIdInt := map[string]int{"UserId": int(userId)}


	db.Close()

	// Return the created user as JSON
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // Set the status code to 200
	json.NewEncoder(w).Encode(userIdInt) // Return the created user
}


// This functions is for other servers to call to check if the user exists
func checkUserExists(w http.ResponseWriter, r *http.Request) {

	type CheckUserRequest struct {
		Email 			string `json:"Email"`
	}

	// Extract `userId` from URL variables
	vars := mux.Vars(r)
	userIdStr, idExists := vars["userId"]

	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/user_db")
	if err != nil {
		fmt.Println("Error connecting to the database")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer db.Close()

	// Logic for checking user by ID
	if idExists && userIdStr != "" {
		// Convert `userId` to integer
		userId, err := strconv.Atoi(userIdStr)
		if err != nil {
			fmt.Println("Invalid userId format")
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid userId"})
			return
		}

		// Check user by ID in the database
		var checkQuery = "SELECT * FROM User WHERE UserId = ?"
		row := db.QueryRow(checkQuery, userId)
		var user User
		err = row.Scan(&user.UserId, &user.Email, &user.FirstName, &user.LastName, &user.MembershipTier)
		if err != nil {
			if err == sql.ErrNoRows {
				// User does not exist
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusNotFound) // 404 Not Found
				json.NewEncoder(w).Encode(ErrorResponse{Message: "User not found"})
				return
			}
			// Other errors
			fmt.Println("Error querying the database:", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}

		// User exists, send the user object as JSON
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK) // 200 OK
		json.NewEncoder(w).Encode(user)
		return
	}

	// Logic for checking user by email (if ID is not provided)
	
	// Read the request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request body"})
		return
	}


	// Parse the request body into `CheckUserRequest`
	var userRequest CheckUserRequest
	err = json.Unmarshal(jsonByte, &userRequest)
	if err != nil {
		fmt.Println("Error unmarshalling JSON", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // 400 Bad Request
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid request format"})
		return
	}

	// Check user by email in the database
	var checkQueryByEmail = "SELECT * FROM User WHERE Email = ?"
	row := db.QueryRow(checkQueryByEmail, userRequest.Email)
	var user User
	err = row.Scan(&user.UserId, &user.Email, &user.FirstName, &user.LastName, &user.MembershipTier)
	if err != nil {
		if err == sql.ErrNoRows {
			// User does not exist
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound) // 404 Not Found
			json.NewEncoder(w).Encode(ErrorResponse{Message: "User not found"})
			return
		}
		// Other errors
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
	}
	// User exists
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK) // 200 OK
	json.NewEncoder(w).Encode(user)
}


func updateUser(w http.ResponseWriter, r *http.Request) {

	type UpdateUserRequest struct {
		Email 			string `json:"Email"`
		FirstName 		string `json:"FirstName"`
		LastName 		string `json:"LastName"`
	}

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
		fmt.Println("Error unmarshalling JSON", err)

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

	// Update the user in the user database
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

	// Call the authentication service to update the user's email
	authURL := "http://localhost:5000/authenticate/update-email/" + strconv.Itoa(userId)
	authReqBody, err := json.Marshal(map[string]string{"Email": user.Email})
	if err != nil {
		fmt.Println("Error marshalling JSON for authentication service")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Create a new PUT request
	req, err := http.NewRequest(http.MethodPut, authURL, bytes.NewBuffer(authReqBody))
	if err != nil {
		fmt.Println("Error creating PUT request", err)
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the PUT request using http.Client
	client := &http.Client{}
	authRes, err := client.Do(req)
	if err != nil {
		fmt.Println("Error calling authentication service", err)
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer authRes.Body.Close()

	fmt.Println("Response from authentication service:", authRes.Status)
	// Check the status code of the response
	if authRes.StatusCode != http.StatusOK {
		fmt.Println("Error updating email in authentication service")
		w.WriteHeader(http.StatusInternalServerError) // Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return success response
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"Message": "Account updated successfully"})
}

func getUser(w http.ResponseWriter, r *http.Request) {
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