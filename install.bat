@echo off
SETLOCAL

REM Define GOOS and GOARCH based on the current system
SET GOOS=windows
SET GOARCH=amd64

REM Function to compile the program
CALL :COMPILE

REM Add the compiled binary to PATH for Windows
CALL :INSTALL_WINDOWS

GOTO :EOF

:COMPILE
echo Compiling for %GOOS%-%GOARCH%
go build -o main.exe main.go
GOTO :EOF

:INSTALL_WINDOWS
echo Installing for Windows...

REM Check if a specific directory for binaries exists, if not create it
IF NOT EXIST "%USERPROFILE%\bin" (
    mkdir "%USERPROFILE%\bin"
)

REM Move the binary
move /Y main.exe "%USERPROFILE%\bin\goanime.exe"

REM Use PowerShell to add the directory to the system PATH
powershell -Command "$env:Path += ';%USERPROFILE%\bin'"

echo Installation complete.
GOTO :EOF
