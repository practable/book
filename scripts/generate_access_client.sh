#!/bin/bash
rm -rf ../internal/ac/models
rm -rf ../internal/ac/restapi
swagger generate client -t ../internal/ac -f ../api/access.yml -A ac
# add in patches
cp ../patch/internal/ac/models/pretty.go.tmpl ../internal/ac/models/pretty.go
go mod tidy
