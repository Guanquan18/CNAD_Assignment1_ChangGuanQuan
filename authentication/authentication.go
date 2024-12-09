package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
    "strings"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	UserId 			int
    Email 			string
	PasswordHash 	string
}

type ErrorResponse struct {
	Message string `json:"Message"`
}



func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/authenticate/login", loginUser).Methods("POST")
	router.HandleFunc("/authenticate/signup", signUpUser).Methods("POST")
    router.HandleFunc("/authenticate/update-email/{UserId}", updateEmail).Methods("PUT")

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

func updateEmail(w http.ResponseWriter, r *http.Request) {
    type UpdateEmailRequest struct {
        Email string `json:"Email"`
    }

    // Get the UserId from the URL
    params := mux.Vars(r)
    userId := params["UserId"]

    // Read the request body
    jsonByte, err := io.ReadAll(r.Body)
    if err != nil {
        fmt.Println("Error reading request body")

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }

    // Unmarshal the JSON byte array into an UpdateEmailRequest struct
    var updateEmailRequest UpdateEmailRequest
    err = json.Unmarshal(jsonByte, &updateEmailRequest)
    if err != nil {
        fmt.Println("Error unmarshalling JSON", err)

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }


    // Connect to the database
    db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/authentication_db")
    if err != nil {
        fmt.Println(err.Error())

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }

    // Update the email in the database
    var query string = "UPDATE Authentication SET Email = ? WHERE UserId = ?"
    _, err = db.Exec(query, updateEmailRequest.Email, userId)
    if err != nil {
        fmt.Println(err.Error())

        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }
    defer db.Close()
    
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(updateEmailRequest)
}


func signUpUser(w http.ResponseWriter, r *http.Request) {
    type SignUpRequest struct {
        Email     string `json:"email"`
        FirstName string `json:"firstName"`
        LastName  string `json:"lastName"`
        Password  string `json:"password"`
    }

    // Read request body
    jsonByte, err := io.ReadAll(r.Body)
    if err != nil {
        fmt.Println("Error reading request body")
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }

    // Unmarshal the JSON byte array into a SignUpRequest struct
    var user SignUpRequest
    err = json.Unmarshal(jsonByte, &user)
    if err != nil {
        fmt.Println("Error unmarshalling JSON", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }

    // Call the user service to check if the email exists
    externalServiceURL := "http://localhost:5001/account/check-user"
    checkUserPayload := map[string]string{"Email": user.Email}

    // Convert the payload to JSON
    payload, err := json.Marshal(checkUserPayload)
    if err != nil {
        fmt.Println("Error marshalling payload:", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }

    // Make a POST request to the external service
    resp, err := http.Post(externalServiceURL, "application/json", bytes.NewBuffer(payload))
    if err != nil {
        fmt.Println("Error calling external service:", err)
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
        json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
        return
    }
    defer resp.Body.Close()

    // Check the response from the external service
    if resp.StatusCode == http.StatusOK {

        // If user exists, return conflict
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusConflict) // 409 Conflict
        json.NewEncoder(w).Encode(ErrorResponse{Message: "User already exists"})
        return

    } else if resp.StatusCode == http.StatusNotFound {

        // Insert the user into the database by calling another server user.go
        externalServiceURL := "http://localhost:5001/account/create-user"
        createUserPayload := map[string]string{"Email": user.Email, "FirstName": user.FirstName, "LastName": user.LastName}

        // Convert the payload to JSON
        payload, err := json.Marshal(createUserPayload)
        if err != nil {
            fmt.Println("Error marshalling payload:", err)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
            json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
            return
        }

        // Make a POST request to the external service
        resp, err := http.Post(externalServiceURL, "application/json", bytes.NewBuffer(payload))
        if err != nil {
            fmt.Println("Error calling external service:", err)
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
            json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
            return
        }
        defer resp.Body.Close()

        // Check the response from the external service
        if resp.StatusCode == http.StatusOK {
            type UserResponse struct {
                UserId int `json:"UserId"`
            }

            // Read the response body
            jsonByte, err := io.ReadAll(resp.Body)
            if err != nil {
                fmt.Println("Error reading response body")
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
                json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
                return
            }

            // Unmarshal the JSON byte array into a UserResponse struct
            var userResponse UserResponse
            err = json.Unmarshal(jsonByte, &userResponse)
            if err != nil {
                fmt.Println("Error unmarshalling JSON", err)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
                json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
                return
            }

            // Hash the password before storing it
            passwordHash := hashPassword(user.Password)

            // If user doesn't exist, proceed with signup logic
            db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/authentication_db")
            if err != nil {
                fmt.Println("Error connecting to the database:", err)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
                json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
                return
            }
            defer db.Close()

            // Insert the user into the database
            var query string = "INSERT INTO Authentication (UserId, Email, PasswordHash) VALUES (?, ?, ?)"
            _, err = db.Exec(query, userResponse.UserId, user.Email, passwordHash)
            if err != nil {
                fmt.Println("Error inserting user into database:", err)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError) // 500 Internal Server Error
                json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
                return
            }

            // Return the created user
            user := User{UserId: userResponse.UserId, Email: user.Email, PasswordHash: passwordHash}

            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusCreated) // 201 Created
            json.NewEncoder(w).Encode(user)   // Return the created user
            return
        }
    }

    // Handle other statuses
	fmt.Println("Error calling external service:", resp.StatusCode)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
}

// Function to hash the password
func hashPassword(password string) string {
    hash, _ := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
    return string(hash)
}

func comparePassword(password string, hash string) bool {
    hash = strings.TrimSpace(hash) // Ensure the hash has no extra spaces or newlines
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    if err != nil {
        fmt.Println("Password comparison failed:", err)
        return false
    }
    return true
}



func loginUser(w http.ResponseWriter, r *http.Request) {

	type LoginRequest struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Read request body
	jsonByte, err := io.ReadAll(r.Body)
	if err != nil {
		fmt.Println("Error reading request body")

		w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}


	// Unmarshal the JSON byte array into a LoginRequest struct
	var loginRequest LoginRequest
	err = json.Unmarshal(jsonByte, &loginRequest)
	if err != nil {
		fmt.Println("Error unmarshalling JSON", err)

		w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}


	// Connect to the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/authentication_db")
	if err != nil {
		fmt.Println(err.Error())

		w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Check if the user exists in the database
	var query string = "SELECT * FROM Authentication WHERE Email = ?"
	results, err := db.Query(query, loginRequest.Email)
	if err != nil {
		fmt.Println(err.Error())

		w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
		w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Create a slice to store the data and check if the user exists
	var user User
	if results.Next() {
		err = results.Scan(&user.UserId, &user.Email, &user.PasswordHash)
		if err != nil {
			fmt.Println(err.Error())

			w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
			w.WriteHeader(http.StatusInternalServerError)	// Set the status code to 500
			json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
			return
		}
	} else {
		fmt.Println("User not found")
		w.WriteHeader(http.StatusNotFound)	// Set the status code to 404
		return
	}
	defer db.Close()

    // Check if the password matches
    passwordHash := strings.TrimSpace(user.PasswordHash) // Sanitize the hash
    if !comparePassword(loginRequest.Password, passwordHash) {
        fmt.Println("Password does not match")
        w.WriteHeader(http.StatusUnauthorized) // 401 Unauthorized
        return
    }

	w.Header().Set("Content-Type", "application/json") // Set the content type to JSON
	w.WriteHeader(http.StatusOK)	// Set the status code to 200
	json.NewEncoder(w).Encode(user)	// Encode the user into JSON and write it to the response
}