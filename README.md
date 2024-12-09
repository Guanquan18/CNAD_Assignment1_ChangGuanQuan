# CNAD_Assignment1_ChangGuanQuan
 # Electric Car Sharing System

## Table of Contents

1. [Design Considerations](#design-considerations)
2. [Architecture Diagram](#architecture-diagram)
3. [API Documentation](#api-documentation)
4. [Setup Instructions](#setup-instructions)

---

## Design Considerations

### Microservices Design

### **Modularization**
- The system is divided into distinct microservices to improve scalability, maintainability, and fault isolation. Each service is designed with a single responsibility principle, making it easier to develop, test, and deploy independently.

   - **User Service**: Manages user registration, membership tiers, and profile updates.
   - **Reservation Service**: Handles vehicle availability, booking, modifications, and cancellations.
   - **Vehicle Service**: Manages vehicle details and availability
   - **Billing Service**: Computes rental costs based on membership tiers, promotions, and rental duration. It handles payment processing and invoice generation.
   - **Authentication Service**: Provides secure authentication and authorization for user registration, utilizing encrypted passwords.

### **Inter-Service Communication**
- All microservices communicate via **RESTful APIs**:
   
   - Promotes simplicity and ensures clear documentation.
   - REST APIs facilitate interoperability, allowing other systems to interact with the services seamlessly.
   - Each service defines its endpoints and maintains a consistent interface for integration.

### **Database Design**
- A **relational database** is used for all services to ensure structured and normalized data storage. Each service interacts with its own database:
   - **User Database**: Stores user details, membership tiers, and profiles.
   - **Vehicle Database**: Manages vehicle inventory and status.
   - **Reservation Database**: Tracks reservation details and policies.
   - **Billing Database**: Stores invoices, promotions, and transaction records.

### **Scalability**
- Microservices are independently scalable based on traffic and computational needs:
   - The Reservation and Billing Services may scale up during peak booking times or promotional periods.
   - The Vehicle Service may scale to accommodate vehicle availability searching.

### **Security** 
- Authentication and authorization are implemented using hashed passwords, and sensitive data is protected during transmission and storage.

---


## Architecture Diagram

![CNAD Assignment 1 Diagram drawio](https://github.com/user-attachments/assets/72eee9db-1fb8-4ad0-a942-b3d55935e33c)

---


# **Authentication Service API Documentation**
## **API Endpoints**

### 1. **User Login**
- **Endpoint**: `POST /authenticate/login`
- **Description**: Authenticate a user by verifying their email and password.
- **Request Body**: 
   ```json
   {
     "email": "user@example.com",
     "password": "userpassword"
   }
   ```
- **Response**:
  - 200 OK
   ```json
   {
     "UserId": 1,
     "Email": "user@example.com",
     "PasswordHash": "$2a$10$..."
   }
   ```
   - 401 Unauthorized: Password mismatch.
   - 404 Not Found: User does not exist.

### 2. **User Signup**
- **Endpoint**: `POST /authenticate/signup`
- **Description**: Register a new user. If the email already exists, the service returns a conflict status.
- **Request Body**:
  ```json
  {
    "email": "user@example.com",
    "firstName": "John",
    "lastName": "Doe",
    "password": "securepassword"
  }
  ```
- **Response**:
  - 201 Created
  ```json
  {
    "UserId": 2,
    "Email": "user@example.com",
    "PasswordHash": "$2a$10$..."
  }
  ```
  - 409 Conflict: User already exists.


### 3. **Update User Email**
- **Endpoint**: `PUT /authenticate/update-email/{UserId}`
- **Description**: Update the email address associated with a user.
- **Path Parameter**:
  - `UserId:` The ID of the user whose email is being updated.
Request Body:
  ```json
  {
    "Email": "newemail@example.com"
  }
  ```
- **Response**:
  - **200 OK**
  ```json
  {
    "Email": "newemail@example.com"
  }
  ```
  - **500 Internal Server Error**: Database or server-related errors.

---

# **User Service API Documentation**
## **API Endpoints**

### 1. **Get User Details**
- **Endpoint**: `GET /account/{userId}`
- **Description**: Retrieve the details of a user by their ID.
- **Path Parameters**:
  - `userId` (integer): The ID of the user.
- **Response**:
  - **200 OK**:
    ```json
    {
      "UserId": 1,
      "Email": "user@example.com",
      "FirstName": "John",
      "LastName": "Doe",
      "MembershipTier": "Basic"
    }
    ```
  - **404 Not Found**: User does not exist.
  - **500 Internal Server Error**: Database or server-related errors.

---

### 2. **Update User Details**
- **Endpoint**: `PUT /account/update/{userId}`
- **Description**: Update the details of a user by their ID.
- **Path Parameters**:
  - `userId` (integer): The ID of the user to update.
- **Request Body**:
  ```json
  {
    "Email": "newemail@example.com",
    "FirstName": "Jane",
    "LastName": "Doe"
  }
  ```
- **Response**:
  - **200 OK**: Account updated successfully.
    ```json
    {
      "Message": "Account updated successfully"
    }
    ```
  - **400 Bad Request**: Invalid `userId` format.
  - **500 Internal Server Error**: Database or server-related errors.

---

### 3. **Create User**
- **Endpoint**: `POST /account/create-user`
- **Description**: Create a new user account.
- **Request Body**:
  ```json
  {
    "Email": "user@example.com",
    "FirstName": "John",
    "LastName": "Doe"
  }
  ```
- **Response**:
  - **200 OK**: User created successfully.
    ```json
    {
      "UserId": 1
    }
    ```
  - **400 Bad Request**: Invalid request format.
  - **500 Internal Server Error**: Database or server-related errors.

---

### 4. **Check User Existence**
- **Endpoint**: `POST /account/check-user/{userId}` or `POST /account/check-user`
- **Description**: Verify if a user exists by ID or email.
- **Path Parameters** (optional):
  - `userId` (integer): The ID of the user.
- **Request Body** (for email-based check):
  ```json
  {
    "Email": "user@example.com"
  }
  ```
- **Response**:
  - **200 OK**: User exists.
    ```json
    {
      "UserId": 1,
      "Email": "user@example.com",
      "FirstName": "John",
      "LastName": "Doe",
      "MembershipTier": "Basic"
    }
    ```
  - **404 Not Found**: User does not exist.
  - **500 Internal Server Error**: Database or server-related errors.

---

### 5. **Check Reservation Limit**
- **Endpoint**: `GET /account/check-reservation-limit/{userId}`
- **Description**: Check if a user has exceeded their reservation limit.
- **Path Parameters**:
  - `userId` (integer): The ID of the user.
- **Response**:
  - **200 OK**: Reservation limit not exceeded.
    ```json
    {
      "Message": "Reservation limit is not exceeded"
    }
    ```
  - **406 Not Acceptable**: Reservation limit exceeded.
    ```json
    {
      "Message": "Reservation limit is exceeded"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 6. **Get Membership Discount**
- **Endpoint**: `GET /account/membership-discount/{userId}`
- **Description**: Retrieve the membership discount for a user.
- **Path Parameters**:
  - `userId` (integer): The ID of the user.
- **Response**:
  - **200 OK**:
    ```json
    {
      "MembershipTier": "Premium",
      "ReducedHourlyRate": 15.0
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---
# Vehicle Service API Documentation
## API Endpoints

### 1. **Search Vehicles**
- **Endpoint**: `GET /vehicle`
- **Description**: Retrieve a list of available vehicles for a specific time range.
- **Query Parameters**:
  - `startDate` (string, required): The start date for the vehicle reservation.
  - `endDate` (string, required): The end date for the vehicle reservation.
- **Response**:
  - **200 OK**: List of available vehicles.
    ```json
    [
      {
        "vehicleId": 1,
        "vehicleType": "Sedan",
        "make": "Toyota",
        "model": "Corolla",
        "year": 2020,
        "location": "Downtown",
        "hourlyRate": 15.0,
        "availability": true
      }
    ]
    ```
  - **400 Bad Request**: Missing or invalid query parameters.
    ```json
    {
      "Message": "Missing time range parameters"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 2. **Get Vehicle Details**
- **Endpoint**: `GET /vehicle/get-vehicle/{vehicleId}`
- **Description**: Retrieve the details of a specific vehicle by its ID.
- **Path Parameters**:
  - `vehicleId` (integer): The ID of the vehicle.
- **Response**:
  - **200 OK**: Vehicle details.
    ```json
    {
      "Make": "Toyota",
      "Model": "Corolla",
      "Location": "Downtown"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---
# Reservation Service API Documentation

## API Endpoints

### 1. **Search Reservations**
- **Endpoint**: `GET /reservation`
- **Description**: Search for reservations within a specific date range.
- **Query Parameters**:
  - `startDate` (string, required): The start date for the search.
  - `endDate` (string, required): The end date for the search.
- **Response**:
  - **200 OK**: List of reservations.
    ```json
    [
      {
        "reservationId": 1,
        "vehicleId": 2,
        "userId": 3,
        "startDate": "2023-12-01",
        "endDate": "2023-12-02",
        "totalCost": 150.0,
        "status": "Confirmed"
      }
    ]
    ```
  - **400 Bad Request**: Missing or invalid date range parameters.
    ```json
    {
      "Message": "Missing date range parameters"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 2. **Create Reservation**
- **Endpoint**: `POST /reservation/make-reservation`
- **Description**: Create a new reservation.
- **Request Body**:
  ```json
  {
    "UserId": 3,
    "VehicleId": 2,
    "StartDate": "2023-12-01",
    "EndDate": "2023-12-02",
    "TotalCost": 150.0
  }
  ```
- **Response**:
  - **201 Created**: Reservation created successfully.
    ```json
    {
      "ReservationId": 1
    }
    ```
  - **400 Bad Request**: Invalid request format.
  - **500 Internal Server Error**: Database or server-related errors.

---

### 3. **Get User Reservations**
- **Endpoint**: `GET /reservation/{type}/{userId}`
- **Description**: Retrieve a user's upcoming or past reservations.
- **Path Parameters**:
  - `type` (string, required): Either `upcoming` or `history`.
  - `userId` (integer, required): The ID of the user.
- **Response**:
  - **200 OK**: List of reservations with vehicle details.
    ```json
    [
      {
        "reservation": {
          "reservationId": 1,
          "vehicleId": 2,
          "userId": 3,
          "startDate": "2023-12-01",
          "endDate": "2023-12-02",
          "totalCost": 150.0,
          "status": "Confirmed"
        },
        "vehicle": {
          "make": "Toyota",
          "model": "Corolla",
          "location": "Downtown"
        }
      }
    ]
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 4. **Cancel Reservation**
- **Endpoint**: `PUT /reservation/cancel/{reservationId}`
- **Description**: Cancel a reservation by its ID.
- **Path Parameters**:
  - `reservationId` (integer, required): The ID of the reservation to cancel.
- **Response**:
  - **200 OK**: Reservation cancelled successfully.
    ```json
    {
      "message": "Reservation cancelled successfully"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 5. **Update Reservation**
- **Endpoint**: `PUT /reservation/update/{reservationId}`
- **Description**: Update the details of an existing reservation.
- **Path Parameters**:
  - `reservationId` (integer, required): The ID of the reservation to update.
- **Request Body**:
  ```json
  {
    "StartDate": "2023-12-03",
    "EndDate": "2023-12-04",
    "VehicleId": 3
  }
  ```
- **Response**:
  - **200 OK**: Reservation updated successfully.
  - **409 Conflict**: Another reservation exists for the same vehicle and time period.
    ```json
    {
      "Message": "Another reservation exists for the same vehicle and time period"
    }
    ```
  - **500 Internal Server Error**: Database or server-related errors.

---

### 6. **Check Reservation Limit**
- **Endpoint**: `POST /reservation/check-reservation-limit/{userId}`
- **Description**: Verify if a user has reached their reservation limit.
- **Path Parameters**:
  - `userId` (integer, required): The ID of the user.
- **Request Body**:
  ```json
  {
    "BookingLimit": 3
  }
  ```
- **Response**:
  - **200 OK**: Reservation limit not exceeded.
  - **406 Not Acceptable**: Reservation limit exceeded.
  - **500 Internal Server Error**: Database or server-related errors.

---

# Billing Service API Documentation

## API Endpoints

### 1. **Get Membership Discount**
- **Endpoint**: `GET /billing/membership-discount/{userId}`
- **Description**: Retrieve the membership discount for a specific user.
- **Path Parameters**:
  - `userId` (integer, required): The ID of the user.
- **Response**:
  - **200 OK**:
    ```json
    {
      "MembershipTier": "Premium",
      "ReducedHourlyRate": 10.0
    }
    ```
  - **500 Internal Server Error**: Error fetching membership discount.

---

### 2. **Get Discount Code Details**
- **Endpoint**: `POST /billing/discount-code`
- **Description**: Validate and retrieve details of a discount code.
- **Request Body**:
  ```json
  {
    "code": "PROMO123"
  }
  ```
- **Response**:
  - **200 OK**: Valid discount code.
    ```json
    {
      "PromotionId": 1,
      "DiscountValue": 15.0
    }
    ```
  - **404 Not Found**: Discount code not found.
    ```json
    {
      "Message": "Discount code not found"
    }
    ```
  - **500 Internal Server Error**: Error querying the discount code.

---

### 3. **Make Payment**
- **Endpoint**: `POST /billing/make-payment`
- **Description**: Process a payment and generate an invoice.
- **Request Body**:
  ```json
  {
    "userId": 1,
    "userEmail": "user@example.com",
    "vehicleId": 2,
    "vehicleMakeModel": "Toyota Corolla",
    "vehicleLocation": "Downtown",
    "totalAmount": 150.0,
    "membershipTier": "Premium",
    "promotionId": 1,
    "promotionCode": "PROMO123",
    "finalAmount": 135.0,
    "cardNumber": "1234-5678-9012-3456",
    "startDate": "2023-12-01",
    "endDate": "2023-12-02"
  }
  ```
- **Response**:
  - **201 Created**: Payment processed successfully.
    ```json
    {
      "BillingId": 101
    }
    ```
  - **400 Bad Request**: Invalid request format.
  - **500 Internal Server Error**: Error processing payment.

---

## Setup Instructions

## Setup Instructions

### Starting All Services Simultaneously

1. **Clone the Repository**:
   ```bash
   git clone https://github.com/your-username/electric-car-sharing.git
   ```

2. **Database Setup**:
   - Import the provided SQL script to set up the database.
   - Update the database credentials in the `.env` file or configuration files.

3. **Install Dependencies**:
   ```bash
   go mod tidy
   ```

4. **Run All Services**:
   Use the provided batch script to start all services simultaneously:
   ```bash
   ./run_servers.bat
   ```

5. **API Testing**:
   - Use Postman or cURL to test the endpoints.
   - Ensure all services are running for testing inter-service communication.

---

## Notes

- Ensure all required ports are available.
- Refer to individual service logs for debugging.
- Use the `README.md` for each service to understand its specific setup.
