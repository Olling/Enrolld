#/bin/bash

counter=40000
path=$1

if [ "$path" == "" ]
then
	echo "Exiting - Missing path"
	exit 1
fi

until [ $counter -lt 1 ]
do
	server="{
        \"ServerID\": \"Server$counter\",
        \"IP\": \"1.2.3.4\",
        \"LastSeen\": \"2020-02-20 20:00:00.423709525 +0100 CET m=+1935960.050243285\",
        \"Inventories\": [
                \"inventory1\",
                \"inventory2\"
        ],
        \"AnsibleProperties\": {
                \"domain\": \"Olling.local\",
                \"environment\": \"internal\"
        }
}"

	echo -e "$server" > "$path/server$counter"

	let counter=counter-1
done