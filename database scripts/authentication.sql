/* MySQL script to create the Authentication database and table */

/* Create the database if it does not exist */
CREATE DATABASE IF NOT EXISTS authentication_db;

USE authentication_db;

/* Drop the Authentication table if it exists */
DROP TABLE IF EXISTS Authentication;

/* Create the Authentication table */
CREATE TABLE Authentication (
    UserId INT PRIMARY KEY,
    Email VARCHAR(100) NOT NULL UNIQUE,
    PasswordHash VARCHAR(200) NOT NULL
);

/* Insert some random data into the table */
INSERT INTO Authentication (UserId, Email, PasswordHash)
VALUES
(1,'john.doe@example.com', '$2a$10$g6TNDp4gQWsvnMtlf6lNqeqlZ106VGexeiVvB7oypOdeh1gUG/qD2'),
(2,'jane.smith@example.com', '$2a$10$AERLrkCcu843im/hwoYODe447KGw5n1kFLo1f.50V1seT3hRY5x/W'),
(3,'sam.jones@example.com', '$2a$10$ourt6ic40o8N0hJdOWgba.dCUrx2CdXzl6XSnObZ.IvLxh9/pp3UC'),
(4,'emily.brown@example.com', '$2a$10$c89LKa8JEG47fN4LF78YW.cP118jKG/WnbJQ8QnC1wXiNDLoB3Yyi'),
(5,'michael.white@example.com', '$2a$10$EJSWtCm9qWcgrnWD10GLBujQ3YD.2maaRi.pmW87BfhZWiEgwAxlq');

use authentication_db;
SELECT * FROM Authentication;