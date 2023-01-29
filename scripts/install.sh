#!/bin/bash
version=$(./build.sh)
sudo cp ../cmd/book/book /usr/local/bin/book
echo "$version installed to /usr/local/bin/book"
