#!/bin/bash

# Build-Script für MLCProxy
echo "Building MLCProxy..."

# Erstelle dist-Ordner und Unterordner
DIST_PATH="dist"
STATIC_PATH="$DIST_PATH/static"
mkdir -p "$STATIC_PATH"

# Baue das Programm
echo "Compiling..."
go build -o "$DIST_PATH/mlcproxy" cmd/proxy/main.go
if [ $? -ne 0 ]; then
    echo "Build failed!"
    exit 1
fi

# Kopiere statische Dateien
echo "Copying static files..."
cp -r internal/stats/static/* "$STATIC_PATH"
cp config.ini "$DIST_PATH"
cp LICENSE "$DIST_PATH"

echo "Build complete! Files are in $DIST_PATH/"
echo "To run: cd $DIST_PATH && ./mlcproxy"
