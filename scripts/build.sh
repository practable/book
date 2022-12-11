#!/bin/bash
export GOOS=linux
now=$(date +'%Y-%m-%d_%T') #no spaces to prevent quotation issues in build command
(cd ../cmd/book; go build -ldflags "-X 'github.com/timdrysdale/interval/cmd/book/cmd.Version=`git describe`' -X 'github.com/timdrysdale/interval/cmd/book/cmd.BuildTime=$now'"; ./book version)
