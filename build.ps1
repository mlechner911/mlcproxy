# Build-Skript für MLCProxy
Write-Host "Building MLCProxy..." -ForegroundColor Green

# Erstelle dist-Ordner und Unterordner
$distPath = "dist"
$staticPath = "$distPath\static"
New-Item -ItemType Directory -Force -Path $distPath | Out-Null
New-Item -ItemType Directory -Force -Path $staticPath | Out-Null

# Baue das Programm
Write-Host "Compiling..." -ForegroundColor Yellow
go build -o "$distPath\mlcproxy.exe" cmd\proxy\main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

# Kopiere statische Dateien
Write-Host "Copying static files..." -ForegroundColor Yellow
Copy-Item "internal\stats\static\*" -Destination "$staticPath" -Recurse -Force
Copy-Item "config.ini" -Destination "$distPath" -Force
Copy-Item "LICENSE" -Destination "$distPath" -Force
Copy-Item "README.md" -Destination "$distPath" -Force
Copy-Item "README.de.md" -Destination "$distPath" -Force

Write-Host "`nBuild completed successfully!" -ForegroundColor Green
Write-Host "You can now run MLCProxy from the 'dist' directory."
