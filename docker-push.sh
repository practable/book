#!/bin/bash
#echo "images for v0.0.1 have alreay been pushed - update push script!"
#exit
docker tag book_book:latest practable/book:0.0.1-alpine
echo "You probably need to do $docker login -u practable #enter password for account admin@practable.io at prompt"
docker push practable/book:0.0.1-alpine
