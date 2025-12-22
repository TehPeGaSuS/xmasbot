#!/bin/bash
# Builds the bot for some platforms
# x86_64
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o xmasbot-linux-amd64

# ARM64 (Pi 4, cloud ARM)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o xmasbot-linux-arm64

# ARMv7 (older Pi)
CGO_ENABLED=0 GOOS=linux GOARCH=arm GOARM=7 go build -o xmasbot-linux-armv7

# Windows 64bits
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -o xmasbot-windows64.exe
