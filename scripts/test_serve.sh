#!/bin/bash

# serve.sh is a script to help with loading
# and resetting the manifest at
# book.practable.io

function freeport(){
 #https://unix.stackexchange.com/questions/55913/whats-the-easiest-way-to-find-an-unused-local-port
 port=$(comm -23 <(seq 49152 65535 | sort) <(ss -Htan | awk '{print $4}' | cut -d':' -f2 | sort -u) | sort -n | head -n 1)
}

freeport

export BOOK_PORT="${port}"
export BOOK_LOGINTIME=1h
export BOOK_FQDN=localhost
export BOOK_ADMIN_SECRET=replace-me-with-some-long-secret
export BOOK_RELAY_SECRET=replace-me-with-another-long-secret
export BOOK_LOG_FILE=stdout
export BOOK_LOG_LEVEL=debug
export BOOK_LOG_FORMAT=text

../cmd/book/book serve &
book_pid=$!


# pad base64URL encoded to base64
# from https://gist.github.com/angelo-v/e0208a18d455e2e6ea3c40ad637aac53
paddit() {
  input=$1
  l=`echo -n $input | wc -c`
  while [ `expr $l % 4` -ne 0 ]
  do
    input="${input}="
    l=`echo -n $input | wc -c`
  done
  echo $input
}

if [ "$BOOK_ADMIN_SECRET" = "" ];
then
	echo 'you must set BOOK_ADMIN_SECRET'
fi

export BOOKCLIENT_HOST="localhost:${port}"
export BOOKCLIENT_SCHEME=http
export BOOKCLIENT_SECRET=$BOOK_ADMIN_SECRET
export BOOKCLIENT_TOKEN_AUD=localhost
export BOOKCLIENT_TOKEN_TTL=5m
export BOOKCLIENT_TOKEN_ADMIN=true
export BOOKCLIENT_TOKEN_SUB=admin
export BOOKCLIENT_TOKEN=$(../cmd/book/book token)
echo "Admin token:"
echo ${BOOKCLIENT_TOKEN}

# read and split the token and do some base64URL translation
read h p s <<< $(echo $BOOKCLIENT_TOKEN | tr [-_] [+/] | sed 's/\./ /g')

h=`paddit $h`
p=`paddit $p`
# assuming we have jq installed
echo $h | base64 -d | jq
echo $p | base64 -d | jq

set | grep BOOKCLIENT


echo "book server at ${BOOKCLIENT_HOST} (testing)"

echo "commands:"
echo "  g: start insecure chrome"
echo "  l: Lock bookings"
echo "  m: replace manifest"
echo "  n: uNlock bookings"
echo "  s: get the status of the poolstore)"


for (( ; ; ))
do
	read -p 'What next? [g/l/m/n/s]:' command
if [ "$command" = "g" ];
then
	mkdir -p ~/tmp/chrome-user
	google-chrome --disable-web-security --user-data-dir="~/tmp/chrome-user" > chrome.log 2>&1 &	
elif [ "$command" = "l" ];
then
	read -p 'Enter lock message:' message
	../cmd/book/book setstatus lock "$message"
elif [ "$command" = "m" ];
then
	echo "NOT IMPLEMENTED"
elif [ "$command" = "n" ];
then
	read -p 'Enter unlock message:' message
	../cmd/book/book setstatus unlock "$message"
elif [ "$command" = "s" ];
then
 	../cmd/book/book getstatus
	
elif [ "$command" = "u" ];
then
	echo "NOT IMPLEMENTED"
	#read -p "Definitely upload [y/N]?" confirm
	#if ([ "$confirm" == "y" ] || [ "$confirm" == "Y" ]  || [ "$confirm" == "yes"  ] );
	#then
	#	export BOOKTOKEN_ADMIN=true
    #	export BOOKUPLOAD_TOKEN=$(book token)
	#	book upload manifest.yaml
	#else
	#	echo "wise choice, aborting"
	#fi
else	
     echo -e "\nUnknown command ${command}."
fi
done

kill book_pid

