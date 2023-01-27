#!/bin/bash
(cd cmd/book ; go build )

export BOOK_FQDN=http://localhost
export BOOK_PORT=4000
export BOOK_ADMIN_SECRET=$(cat ~/secret/book.pat)
export BOOK_RELAY_SECRET=$(cat ~/secret/sessionrelay.pat)
export BOOK_LOG_LEVEL=debug
export BOOK_LOG_FORMAT=text
export BOOK_ACCESS_TOKEN_TTL=1h
export BOOK_CHECK_EVERY=1m
export BOOK_TIDY_EVERY=1h
export BOOK_MIN_USERNAME_LENGTH=6

./cmd/book/book serve

