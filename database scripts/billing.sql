-- Create the Billing database
CREATE DATABASE IF NOT EXISTS billing_db;

USE billing_db;

-- Drop the Billing table if it exists
DROP TABLE IF EXISTS Billing;
DROP TABLE IF EXISTS Promotion;

-- Create the Promotion table
CREATE TABLE Promotion(
    PromotionId INT AUTO_INCREMENT PRIMARY KEY,
    Code VARCHAR(50) NOT NULL UNIQUE,
    Description TEXT,
    DiscountValue DECIMAL(10, 2) NOT NULL,
    StartDate DATE NOT NULL,
    EndDate DATE NOT NULL
);
-- Insert sample data into Promotion table
INSERT INTO Promotion (Code, Description, DiscountValue, StartDate, EndDate)
VALUES
('WELCOME10', '10% off for new users',  10.00, '2024-01-01', '2024-12-31'),
('FLAT50', 'Flat $50 off for orders above $200', 50.00,  '2024-02-01', '2024-06-30'),
('SUMMER15', '15% off during summer sale', 15.00, '2024-06-01', '2024-08-31'),
('PLATINUM20', '20% off for platinum members', 20.00, '2024-01-01', '2024-12-31');


-- Create the Billing table
CREATE TABLE Billing (
    BillingId INT AUTO_INCREMENT PRIMARY KEY,
    ReservationId INT NOT NULL,
    TotalAmount DECIMAL(10, 2) NOT NULL,
    MembershipTier VARCHAR(20),
    PromotionId INT,
    cardNumber char(16),
    FinalAmount DECIMAL(10, 2) NOT NULL, 
    TransactionDate TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (PromotionId) REFERENCES Promotion(PromotionId) ON DELETE SET NULL
);
-- Insert some sample data (optional)
INSERT INTO Billing (ReservationId,TotalAmount, MembershipTier, PromotionId, FinalAmount, cardNumber)
VALUES
(1, 120.00, 'Premium', 1, 108.00, '************1234'),
(2, 150.00, 'Basic', NULL, 150.00, '************1234'),
(3, 200.00, 'VIP', 4, 160.00, '************1234');


-- Test the setup
USE billing_db;
SELECT * FROM Billing;
SELECT * FROM Promotion;