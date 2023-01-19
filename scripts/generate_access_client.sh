#!/bin/bash
rm -rf ../internal/ac/models
rm -rf ../internal/ac/restapi
swagger generate client -t ../internal/ac -f ../api/access.yml -A ac
go mod tidy
