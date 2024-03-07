# Define GOOS and GOARCH based on the current system
$env:GOOS = "windows"
$env:GOARCH = "amd64"

# Function to compile the program
Function Compile {
    Write-Host "Compiling for $($env:GOOS)-$($env:GOARCH)"
    go build -o main.exe .
}

# Function to install the compiled binary for Windows
Function Install-Windows {
    Write-Host "Installing for Windows..."

    # Check if a specific directory for binaries exists, if not create it
    $binDir = "$env:USERPROFILE\bin"
    If (!(Test-Path $binDir)) {
        New-Item -Path $binDir -ItemType Directory
    }

    # Move the binary
    Move-Item -Path main.exe -Destination "$binDir\goanime.exe" -Force

    # Use PowerShell to add the directory to the system PATH
    $env:Path += ";$binDir"

    Write-Host "Installation complete."
}

# Call the functions
Compile
Install-Windows
