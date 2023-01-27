#!/bin/bash
export GOOS=linux
now=$(date +'%Y-%m-%d_%T') #no spaces to prevent quotation issues in build command
(cd ../cmd/book; go build -ldflags "-X 'github.com/practable/book/cmd/book/cmd.Version=`git describe --tags`' -X 'github.com/practable/book/cmd/book/cmd.BuildTime=$now'"; ./book version)
