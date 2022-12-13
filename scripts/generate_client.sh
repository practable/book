#!/bin/bash
rm -rf ../internal/client/models
rm -rf ../internal/client/restapi
swagger generate client -t ../internal/client -f ../api/booking.yml -A client
# add in patches
cp ../patch/internal/client/models/pretty.go ../internal/client/models/
go mod tidy
