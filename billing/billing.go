package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"io"
	"net/http"
	"bytes"
	"os"
	"time"
	"net/smtp"
	"text/template"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

// Define a struct to hold reservation data
type Billing struct {
	BillingId       int       `json:"billingId"`
	ReservationId   int       `json:"reservationId"`
	TotalAmount     float64   `json:"totalAmount"`
	MembershipTier  string    `json:"membershipTier"`
	PromotionId     *int      `json:"promotionId"` // Pointer to handle NULL values if PromotionId can be null
	FinalAmount     float64   `json:"finalAmount"`
	TransactionDate time.Time `json:"transactionDate"`
}


type ErrorResponse struct {
	Message string `json:"Message"`
}

type InvoiceData struct {
	InvoiceNumber    string
	UserId           int
	VehicleId        int
	VehicleMakeModel string
	VehicleLocation  string
	StartDate        string
	EndDate          string
	PromotionCode	*string
	TotalAmount      float64
	FinalAmount      float64
	TransactionDate  string
}

func main() {
	router := mux.NewRouter()

	// Application routes
	router.HandleFunc("/billing/membership-discount/{userId}", getMembershipDiscount).Methods("GET")
	router.HandleFunc("/billing/discount-code", getDiscountCode).Methods("POST")
	router.HandleFunc("/billing/make-payment", makePayment).Methods("POST")

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

	fmt.Println("Listening at port 5004")
	log.Fatal(http.ListenAndServe(":5004", recoveryHandler))
}

func makePayment(w http.ResponseWriter, r *http.Request) {

	defer func() {
        if err := recover(); err != nil {
            http.Error(w, fmt.Sprintf("Server encountered an error: %v", err), http.StatusInternalServerError)
        }
    }()

	// Decode the request body into a struct
	type paymentRequest struct {
		UserId	   			int     `json:"userId"`
		UserEmail 			string  `json:"userEmail"`
		VehicleId  			int     `json:"vehicleId"`
		VehicleMakeModel 	string 	`json:"vehicleMakeModel"`
		VehicleLocation 	string 	`json:"vehicleLocation"`
		TotalAmount 		float64 `json:"totalAmount"`
		MembershipTier 		string 	`json:"membershipTier"`
		PromotionId 		*int 	`json:"promotionId"`
		PromotionCode 		string 	`json:"promotionCode"`
		FinalAmount 		float64 `json:"finalAmount"`
		CardNumber 			string 	`json:"cardNumber"`
		StartDate 			string 	`json:"startDate"`
		EndDate 			string 	`json:"endDate"`
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

	// Unmarshal the JSON byte array into a struct
	var paymentReq paymentRequest
	err = json.Unmarshal(jsonByte, &paymentReq)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid JSON"})
		return
	}

	// Send the make reservation request to the reservation service
	externalServiceURL := `http://localhost:5003/reservation/make-reservation`
	// Convert values into integer or decimals for the specific fields
	type CreateUserPayload struct {
		UserId      int     `json:"UserId"`
		VehicleId   int     `json:"VehicleId"`
		TotalCost   float64 `json:"TotalCost"`
		StartDate   string  `json:"StartDate"`
		EndDate     string  `json:"EndDate"`
	}

	var createUserPayload CreateUserPayload
	createUserPayload.UserId = paymentReq.UserId
	createUserPayload.VehicleId = paymentReq.VehicleId
	createUserPayload.TotalCost = paymentReq.FinalAmount
	createUserPayload.StartDate = paymentReq.StartDate
	createUserPayload.EndDate = paymentReq.EndDate

	// Convert the payload map into a JSON byte array
	payloadByte, err := json.Marshal(createUserPayload)
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Make a POST request to the external service
	resp, err := http.Post(externalServiceURL, "application/json", bytes.NewBuffer(payloadByte))
	if err != nil {
		fmt.Println("Error calling external service:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer resp.Body.Close()

	// Check if the external service returned an error
	if resp.StatusCode != http.StatusCreated {
		fmt.Println("External service returned an error:", resp.Status)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Decode the response body into a struct
	type reservationResponse struct {
		ReservationId int `json:"ReservationId"`
	}

	// Read the response body and decode it into the struct
	var resRes reservationResponse
	err = json.NewDecoder(resp.Body).Decode(&resRes)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Insert the billing record into the database
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/billing_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}
	
	// Insert the billing record into the database
	result, err := db.Exec("INSERT INTO Billing (ReservationId, TotalAmount, MembershipTier, PromotionId, FinalAmount, CardNumber) VALUES (?, ?, ?, ?, ?, ?)", 
	resRes.ReservationId, paymentReq.TotalAmount, paymentReq.MembershipTier, paymentReq.PromotionId, paymentReq.FinalAmount, paymentReq.CardNumber)
	if err != nil {
		fmt.Println("Error inserting billing record into the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Get the auto-generated billing ID
	billingId, err := result.LastInsertId()
	if err != nil {
		fmt.Println("Error getting the billing ID:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return the billing ID as JSON
	type Response struct {
		BillingId int `json:"BillingId"`
	}

	var response Response
	response.BillingId = int(billingId)


	// Generate the invoice data
	invoiceData := InvoiceData{
        InvoiceNumber:   	fmt.Sprintf("INVOICE-%d", billingId),
        UserId:          	paymentReq.UserId,
        VehicleId:       	paymentReq.VehicleId,
		VehicleMakeModel: 	paymentReq.VehicleMakeModel,
		VehicleLocation: 	paymentReq.VehicleLocation,
        StartDate:       	paymentReq.StartDate,
        EndDate:         	paymentReq.EndDate,
		PromotionCode:		func() *string { if paymentReq.PromotionCode == "" { none := "None"; return &none } else { return &paymentReq.PromotionCode } }(),
        TotalAmount:     	paymentReq.TotalAmount,
        FinalAmount:     	paymentReq.FinalAmount,
        TransactionDate: 	time.Now().Format("2006-01-02 15:04:05"),
    }

	// Generate the HTML invoice
	htmlFilePath := "invoice.html"
	err = generateInvoiceHTML(invoiceData, htmlFilePath)
	if err != nil {
		fmt.Println("Error generating HTML invoice:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Failed to generate HTML invoice"})
	}

	// Read the generated HTML file for including it in the email body
	htmlContentBytes, err := os.ReadFile(htmlFilePath)
	if err != nil {
		fmt.Println("Error reading HTML file:", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Failed to read HTML file"})
	}
	htmlContent := string(htmlContentBytes)

	// Send the email with the HTML content in the body and as an attachment
	recipientEmail := paymentReq.UserEmail
	go func() {
		err := sendEmailWithAttachment("Your Invoice", "Please find your invoice below and attached.", recipientEmail, htmlFilePath, htmlContent)
		if err != nil {
			log.Printf("Error sending email: %v\n", err)
		} else {
			log.Println("Email sent successfully.")
		}
	}()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func getDiscountCode(w http.ResponseWriter, r *http.Request) {
	// Decode the request body into a struct
	type discountRequest struct {
		Code string `json:"code"`
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


	// Unmarshal the JSON byte array into a struct
	var discountReq discountRequest
	err = json.Unmarshal(jsonByte, &discountReq)
	if err != nil {
		fmt.Println("Error unmarshalling JSON:", err)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest) // Set the status code to 400
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Invalid JSON"})
		return
	}

	// Execute the discount code query
	db, err := sql.Open("mysql", "root:S10257825A@tcp(127.0.0.1:3306)/billing_db")
	if err != nil {
		fmt.Println("Error connecting to the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error connecting to the database"})
		return
	}

	// Query the database for the discount code
	query := "SELECT PromotionId, DiscountValue FROM Promotion WHERE Code = ? and (EndDate > NOW() and StartDate < NOW());"
	rows, err := db.Query(query, discountReq.Code)
	if err != nil {
		fmt.Println("Error querying the database:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}
	defer rows.Close()

	// Check if the discount code exists
	if !rows.Next() {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Discount code not found"})
		return
	}

	// Read the discount value from the query result
	type Promotion struct {
		PromotionId int     `json:"PromotionId"`
		DiscountValue    float64 `json:"DiscountValue"`
	}

	// Scan the PromotionId and DiscountValue from the query result into struct
	var promotion Promotion
	err = rows.Scan(&promotion.PromotionId, &promotion.DiscountValue)
	if err != nil {
		fmt.Println("Error scanning row:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Internal server error"})
		return
	}

	// Return the discount value as a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(promotion)
}
	
func getMembershipDiscount(w http.ResponseWriter, r *http.Request) {
	// Get the userId from the request URL
	vars := mux.Vars(r)
	userId := vars["userId"]

	// Call another service to get the membership tier and discount
	externalServiceURL := `http://localhost:5001/account/membership-discount/` + userId

	// Make a GET request to the external service
	resp, err := http.Get(externalServiceURL)
	if err != nil {
		fmt.Println("Error calling external service:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error calling external service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("External service returned an error:", resp.Status)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "External service returned an error"})
		return
	}

	// Decode the response body into a struct
	type userResponse struct {
		MembershipTier 			string `json:"MembershipTier"`
		ReducedHourlyRate       float64 `json:"ReducedHourlyRate"`
	}

	// Read the response body and decode it into the struct
	var userRes userResponse
	err = json.NewDecoder(resp.Body).Decode(&userRes)
	if err != nil {
		fmt.Println("Error decoding response body:", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{Message: "Error decoding response body"})
		return
	}

	// Return the membership tier and discount as a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(userRes)
}

// aync Function to generate an invoice
func generateInvoiceHTML(data InvoiceData, filePath string) error {
    const htmlTemplate = `
	<!DOCTYPE html>
	<html>
	<head>
		<title>Invoice</title>
		<style>
			body { font-family: Arial, sans-serif; margin: 20px; }
			.invoice-container { width: 80%; margin: auto; padding: 20px; border: 1px solid #ccc; border-radius: 8px; }
			.header { text-align: center; margin-bottom: 20px; }
			.details { margin: 20px 0; }
			.footer { text-align: center; margin-top: 20px; font-size: 0.9em; color: #888; }
		</style>
	</head>
	<body>
		<div class="invoice-container">
			<div class="header">
				<h1>Invoice: {{.InvoiceNumber}}</h1>
			</div>
			<div class="details">
				<p><strong>User ID:</strong> {{.UserId}}</p>
				<p><strong>Vehicle ID:</strong> {{.VehicleId}}</p>
				<p><strong>Vehicle Make/Model:</strong> {{.VehicleMakeModel}}</p>
				<p><strong>Vehicle Location:</strong> {{.VehicleLocation}}</p>
				<p><strong>Start Date:</strong> {{.StartDate}}</p>
				<p><strong>End Date:</strong> {{.EndDate}}</p>
				<p><strong>Promotion Code:</strong> {{.PromotionCode}}</p>
				<p><strong>Total Amount:</strong> ${{.TotalAmount}}</p>
				<p><strong>Final Amount:</strong> ${{.FinalAmount}}</p>
				<p><strong>Transaction Date:</strong> {{.TransactionDate}}</p>
			</div>
			<div class="footer">
				<p>Thank you for your business!</p>
			</div>
		</div>
	</body>
	</html>
	`

    tmpl, err := template.New("invoice").Parse(htmlTemplate)
    if err != nil {
        return fmt.Errorf("failed to parse template: %v", err)
    }

    file, err := os.Create(filePath)
    if err != nil {
        return fmt.Errorf("failed to create file: %v", err)
    }
    defer file.Close()

    err = tmpl.Execute(file, data)
    if err != nil {
        return fmt.Errorf("failed to execute template: %v", err)
    }
    return nil
}


func sendEmailWithAttachment(subject, body, to, attachmentPath, invoiceHTMLContent string) error {
    log.Println("Starting email sending process...")
    smtpServer := "smtp.gmail.com"
    smtpPort := "587"
    email := "Guanquan18@gmail.com" // Replace with your email
    appPassword := "hdij hmnp fhnz mjkh" // Replace with your app password

    log.Println("Reading attachment...")
    attachment, err := os.ReadFile(attachmentPath)
    if err != nil {
        log.Printf("Failed to read attachment: %v\n", err)
        return fmt.Errorf("failed to read attachment: %w", err)
    }

    log.Println("Preparing email message...")
    from := email
    message := bytes.NewBuffer(nil)
    message.WriteString(fmt.Sprintf("From: %s\r\n", email))
    message.WriteString(fmt.Sprintf("To: %s\r\n", to))
    message.WriteString(fmt.Sprintf("Subject: %s\r\n", subject))
    message.WriteString("MIME-Version: 1.0\r\n")
    message.WriteString("Content-Type: multipart/alternative; boundary=\"boundary\"\r\n")
    message.WriteString("\r\n--boundary\r\n")
    message.WriteString("Content-Type: text/plain; charset=\"utf-8\"\r\n\r\n")
    message.WriteString(body)
    message.WriteString("\r\n--boundary\r\n")
    message.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n\r\n")
    message.WriteString(invoiceHTMLContent)
    message.WriteString("\r\n--boundary\r\n")
    message.WriteString("Content-Type: text/html; charset=\"utf-8\"\r\n")
    message.WriteString("Content-Disposition: attachment; filename=\"invoice.html\"\r\n\r\n")
    message.Write(attachment)
    message.WriteString("\r\n--boundary--")

    log.Println("Connecting to SMTP server...")
    auth := smtp.PlainAuth("", email, appPassword, smtpServer)
    err = smtp.SendMail(smtpServer+":"+smtpPort, auth, from, []string{to}, message.Bytes())
    if err != nil {
        log.Printf("Failed to send email: %v\n", err)
        return fmt.Errorf("failed to send email: %w", err)
    }

    log.Println("Email sent successfully.")
    return nil
}

