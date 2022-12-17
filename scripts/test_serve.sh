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
echo
echo "  0: Get status"
echo "  1: Lock bookings"
echo "  2: Unlock bookings"
echo 
echo "  3: Export bookings"
echo "  4: Replace bookings"
echo 
echo "  5: Export old bookings"
echo "  6: Replace old bookings"
echo 
echo "  7: Check manifest"
echo "  8: Export manifest"
echo "  9: Replace manifest (JSON)"
echo "  a: Replace manifest (YAML)"
echo 
echo "  b: Export users"
echo "  c: start insecure chrome"

for (( ; ; ))
do
	read -p 'What next? ' command
if [ "$command" = "0" ];
then
 	../cmd/book/book status get
elif [ "$command" = "1" ];
then
	read -p 'Enter lock message:' message
	../cmd/book/book status set lock "$message"
elif [ "$command" = "2" ];
then
	read -p 'Enter unlock message:' message
	../cmd/book/book status set unlock "$message"
elif [ "$command" = "3" ];
then
	export BOOKCLIENT_FORMAT=yaml
	../cmd/book/book bookings export
elif [ "$command" = "4" ];
then
	read -p "Definitely replace [y/N]?" confirm
	if ([ "$confirm" == "y" ] || [ "$confirm" == "Y" ]  || [ "$confirm" == "yes"  ] );
	then
		../cmd/book/book bookings replace ../demo/bookings.yaml #boiler plate code doesn't report the error messages (just get pointer values) ... :-(
		#curl --data-binary "@../demo/bookings.yaml"  -X PUT -H "Authorization: ${BOOKCLIENT_TOKEN}" -H "Content-type: text/plain" "${BOOKCLIENT_HOST}/api/v1/admin/bookings" 
	fi

elif [ "$command" = "5" ];
then
	export BOOKCLIENT_FORMAT=yaml
	../cmd/book/book oldbookings export
elif [ "$command" = "6" ];
then
	read -p "Definitely replace [y/N]?" confirm
	if ([ "$confirm" == "y" ] || [ "$confirm" == "Y" ]  || [ "$confirm" == "yes"  ] );
	then	
	    echo "replace old bookings"
	fi	

elif [ "$command" = "7" ];
then
	../cmd/book/book manifest check ../demo/manifest.yaml
elif [ "$command" = "8" ];
then
	export BOOKCLIENT_FORMAT=yaml
	../cmd/book/book manifest export 
elif [ "$command" = "9" ];
then
	read -p "Definitely replace [y/N]?" confirm
	if ([ "$confirm" == "y" ] || [ "$confirm" == "Y" ]  || [ "$confirm" == "yes"  ] );
	then
	        export BOOKCLIENT_FORMAT=json
	    	../cmd/book/book manifest replace ../demo/manifest.json
	fi

elif [ "$command" = "a" ];
then
    		export BOOKCLIENT_FORMAT=yaml
		../cmd/book/book manifest replace ../demo/manifest2.yaml
		
elif [ "$command" = "b" ];
then
    		
	echo "export users"
elif [ "$command" = "c" ];
then	
	mkdir -p ~/tmp/chrome-user
	google-chrome --disable-web-security --user-data-dir="~/tmp/chrome-user" > chrome.log 2>&1 &
else	
     echo -e "\nUnknown command ${command}."
fi
done

kill book_pid



