#!/bin/bash
#rm -rf ../internal/client/models
#rm -rf ../internal/client/restapi
swagger generate client -t ../internal/client -f ../api/booking.yml -A client
go mod tidy
