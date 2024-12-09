/* Create the Reservation Database */
CREATE DATABASE IF NOT EXISTS reservation_db;
USE reservation_db;

/* Drop the Reservations table if it exists */
DROP TABLE IF EXISTS Reservation;

/* Create the Reservations table */
CREATE TABLE Reservation (
    ReservationId INT AUTO_INCREMENT PRIMARY KEY,
    UserId INT NOT NULL,
    VehicleId INT NOT NULL,
    StartDate DATETIME NOT NULL,
    EndDate DATETIME NOT NULL,
    TotalCost DECIMAL(10, 2) NOT NULL,
    Status VARCHAR(20) DEFAULT 'Confirmed'
);

/* Insert sample data into the Reservations table */
INSERT INTO Reservation (UserId, VehicleId, StartDate, EndDate, TotalCost, Status)
VALUES
-- Past Reservations
(1, 1, '2022-12-15 08:00:00', '2022-12-15 20:00:00', 50.00, 'Confirmed'),
(2, 2, '2022-12-10 10:00:00', '2022-12-10 18:00:00', 75.00, 'Confirmed'),
(3, 3, '2023-01-05 09:00:00', '2023-01-05 21:00:00', 100.00, 'Confirmed'),
(4, 4, '2023-02-10 10:00:00', '2023-02-10 20:00:00', 90.00, 'Confirmed'),
(5, 5, '2023-06-15 08:00:00', '2023-06-15 22:00:00', 70.00, 'Confirmed'),
(4, 4, '2023-07-20 10:00:00', '2023-07-20 19:00:00', 80.00, 'Confirmed'),
(3, 5, '2023-09-25 12:00:00', '2023-09-25 18:00:00', 55.00, 'Confirmed'),
(1, 1, '2023-10-01 09:00:00', '2023-10-01 20:00:00', 95.00, 'Confirmed'),

-- Future Reservations
(1, 1, '2025-01-15 08:00:00', '2025-01-15 22:00:00', 45.00, 'Confirmed'),
(2, 2, '2025-03-10 09:00:00', '2025-03-10 21:00:00', 65.00, 'Confirmed'),
(3, 3, '2025-05-20 10:00:00', '2025-05-20 23:00:00', 110.00, 'Confirmed'),
(4, 4, '2025-08-15 11:00:00', '2025-08-15 22:00:00', 150.00, 'Confirmed'),
(5, 5, '2026-02-20 08:00:00', '2026-02-20 20:00:00', 120.00, 'Confirmed'),
(3, 3, '2026-05-10 10:00:00', '2026-05-10 19:00:00', 75.00, 'Confirmed'),
(2, 2, '2026-09-05 12:00:00', '2026-09-05 21:00:00', 95.00, 'Confirmed'),
(1, 1, '2026-10-10 09:00:00', '2026-10-10 22:00:00', 140.00, 'Confirmed'),

-- Canceled Reservations
(1, 1, '2023-03-10 08:00:00', '2023-03-10 18:00:00', 35.00, 'Cancelled'),
(2, 2, '2023-05-15 10:00:00', '2023-05-15 22:00:00', 50.00, 'Cancelled'),
(3, 3, '2024-06-01 08:00:00', '2024-06-01 22:00:00', 100.00, 'Cancelled'),
(4, 4, '2025-01-15 10:00:00', '2025-01-15 23:00:00', 75.00, 'Cancelled');

USE reservation_db;
SELECT * FROM Reservation;