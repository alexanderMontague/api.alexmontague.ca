#!/bin/bash

# This script is used to check and update your GoDaddy DNS server to the IP address of your current internet connection.
# Special thanks to mfox for his ps script
# https://github.com/markafox/GoDaddy_Powershell_DDNS
#
# First go to GoDaddy developer site to create a developer account and get your key and secret
#
# https://developer.godaddy.com/getstarted
# Be aware that there are 2 types of key and secret - one for the test server and one for the production server
# Get a key and secret for the production server
#
#Update the first 4 variables with your information

domain="alexmontague.ca" # your domain
name="api"               # name of A record to update
key=$GODADDY_KEY         # key for godaddy developer API
secret=$GODADDY_SECRET   # secret for godaddy developer API
ttl="600"                # Time to Live min value 600

# set headers and print meta data
headers="Authorization: sso-key $key:$secret"
echo [ `date +"%A, %b %d, %Y %I:%M %p"` ]

result=$(curl -s -X GET -H "$headers" \
 "https://api.godaddy.com/v1/domains/$domain/records/A/$name")

dnsIp=$(echo $result | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b")
echo "dnsIp:" $dnsIp

# Get public ip address there are several websites that can do this.
ret=$(curl -s GET "http://ipinfo.io/json")
currentIp=$(echo $ret | grep -oE "\b([0-9]{1,3}\.){3}[0-9]{1,3}\b")
echo "currentIp:" $currentIp

if [ "$dnsIp" != "$currentIp" ];
 then
	#echo "Ips are not equal"
	request='[{"data":"'$currentIp'","ttl":3600}]'
	# echo $request
	nresult=$(curl -i -s -X PUT \
		-H "$headers" \
 		-H "Content-Type: application/json" \
 		-d $request "https://api.godaddy.com/v1/domains/$domain/records/A/$name")
	echo "Updated GoDaddy DNS"
	echo "Result:"
	echo $nresult
fi

printf "\n"
