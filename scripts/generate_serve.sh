#!/bin/bash
rm -rf ./serve/models
rm -rf ./serve/restapi
swagger generate server -t ../internal/serve -f ../api/booking.yml --exclude-main -A serve
go mod tidy

