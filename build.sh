#!/bin/bash

go generate
#go build
#GOARCH=386 GOOS=linux go build hued.go
GOOS=linux GOARCH=arm GOARM=5 go build hued.go rice-box.go
