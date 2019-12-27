#!/bin/sh


cd cmd

rice embed-go
env GOOS=linux GOARCH=amd64 go build -o $1broadlinkgo-linux-amd64
env GOOS=linux GOARCH=arm GOARM=5 go build -o $1broadlinkgo-linux-arm5-raspi
env GOOS=linux GOARCH=arm GOARM=6 go build -o $1broadlinkgo-linux-arm6-raspi
env GOOS=linux GOARCH=arm GOARM=7 go build -o $1broadlinkgo-linux-arm7-raspi
env GOOS=linux GOARCH=arm64 go build -o $1broadlinkgo-linux-arm8-raspi4
env GOOS=windows GOARCH=amd64 go build -o $1broadlinkgo-windows-amd64.exe
env GOOS=darwin GOARCH=amd64 go build -o $1broadlinkgo-darwin-amd64
