@echo off
echo Starting authentication service...
cd authentication
start cmd /k "go run authentication.go"

echo Starting user service...
cd ..\user
start cmd /k "go run user.go"

echo Both services are running in separate windows.
pause
