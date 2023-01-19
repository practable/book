#!/bin/bash
rm -rf ../internal/ac/models
rm -rf ../internal/ac/restapi
swagger generate client -t ../internal/ac -f ../api/access.yml -A ac
# add in patches
cp ../patch/internal/ac/models/pretty.go ../internal/ac/models/
go mod tidy
