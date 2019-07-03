#!/bin/sh


cd cmd

rice embed-go
env GOOS=linux GOARCH=amd64 go build -o broadlinkgo-linux-amd64
env GOOS=linux GOARCH=arm GOARM=5 go build -o broadlinkgo-linux-arm5-raspi
env GOOS=windows GOARCH=amd64 go build -o broadlinkgo-windows-amd64
env GOOS=darwin GOARCH=amd64 go build -o broadlinkgo-darwin-amd64
