@echo off
echo Starting authentication service...
cd authentication
start cmd /k "go run authentication.go"

echo Starting user service...
cd ..\user
start cmd /k "go run user.go"

echo Starting vehicle service...
cd ..\vehicle
start cmd /k "go run vehicle.go"

echo Starting reservation service...
cd ..\reservation
start cmd /k "go run reservation.go"

echo Starting billing service...
cd ..\billing
start cmd /k "go run billing.go"

echo Services are running in separate windows.
pause
