/* MySQL script to create the Membership table */
CREATE DATABASE IF NOT EXISTS user_db;
USE user_db;

/* Drop the Membership table if it exists */
DROP TABLE IF EXISTS User;
DROP TABLE IF EXISTS Membership;

/* Create the Membership table */
CREATE TABLE Membership (
    MembershipTier VARCHAR(20) PRIMARY KEY, -- This will reference the MembershipTier in the User table
    ReducedHourlyRate DECIMAL(5, 2) NOT NULL, -- Discounted hourly rate (percentage or fixed)
    BookingLimit INT NOT NULL -- Maximum number of bookings allowed per month
);

/* Insert sample data into the Membership table */
INSERT INTO Membership (MembershipTier, ReducedHourlyRate, BookingLimit)
VALUES
('Basic', 0.00, 2), -- No reduction and limited bookings for basic tier
('Premium', 10.00, 4), -- 10% reduced hourly rate and higher booking limit
('VIP', 20.00, 10); -- 20% reduced hourly rate and maximum booking allowance


/* Create the User table */
CREATE TABLE User (
    UserId INT AUTO_INCREMENT PRIMARY KEY,
    Email VARCHAR(100) NOT NULL UNIQUE,
    FirstName VARCHAR(50),
    LastName VARCHAR(50),
    MembershipTier VARCHAR(20),
    CONSTRAINT fk_user_membership
        FOREIGN KEY (MembershipTier) REFERENCES Membership(MembershipTier)
        ON UPDATE CASCADE
        ON DELETE SET NULL
);

/* Insert some random data into the User table */
INSERT INTO User (Email, FirstName, LastName, MembershipTier)
VALUES 
('john.doe@example.com', 'John', 'Doe', 'Premium'),
('jane.smith@example.com', 'Jane', 'Smith', 'Basic'),
('sam.jones@example.com', 'Sam', 'Jones', 'VIP'),
('emily.brown@example.com', 'Emily', 'Brown', 'Premium'),
('michael.white@example.com', 'Michael', 'White', 'Basic');

/* Test the structure */
use user_db;
SELECT * FROM User;
SELECT * FROM Membership;