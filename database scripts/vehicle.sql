/* Create the Vehicle Database */
CREATE DATABASE IF NOT EXISTS vehicle_db;
USE vehicle_db;

/* Drop the Vehicles table if it exists */
DROP TABLE IF EXISTS Vehicle;

/* Create the Vehicles table */
CREATE TABLE Vehicle (
    VehicleId INT AUTO_INCREMENT PRIMARY KEY,
    VehicleType VARCHAR(50) NOT NULL,
    Make VARCHAR(50) NOT NULL,
    Model VARCHAR(50) NOT NULL,
    Year YEAR NOT NULL,
    Location VARCHAR(100) NOT NULL,
    HourlyRate DECIMAL(10, 2) NOT NULL,
    Availability BOOLEAN DEFAULT TRUE
);

/* Insert some sample data into the Vehicles table */
INSERT INTO Vehicle (VehicleType, Make, Model, Year, Location, HourlyRate, Availability)
VALUES
('Sedan', 'Toyota', 'Corolla', 2022, 'City Center', 15.00, TRUE),
('SUV', 'Honda', 'CR-V', 2021, 'Airport', 25.00, TRUE),
('Sedan', 'Tesla', 'Model 3', 2023, 'City Center', 50.00, FALSE),
('Truck', 'Ford', 'F-150', 2020, 'Industrial Area', 35.00, TRUE),
('Hatchback', 'Hyundai', 'i20', 2021, 'Suburbs', 10.00, TRUE);

use vehicle_db;
select * from Vehicle;