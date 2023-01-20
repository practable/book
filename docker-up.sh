#!/bin/bash

cat << EOF > book.env
BOOK_FQDN=http://localhost
BOOK_PORT=4000
BOOK_ADMIN_SECRET=$(cat ~/secret/book.pat)
BOOK_RELAY_SECRET=$(cat ~/secret/sessionrelay.pat)
BOOK_LOG_LEVEL=warn
BOOK_LOG_FORMAT=text
BOOK_ACCESS_TOKEN_TTL=1h
BOOK_TIDY_EVERY=1h
BOOK_MIN_USERNAME_LENGTH=6
EOF

docker-compose up

