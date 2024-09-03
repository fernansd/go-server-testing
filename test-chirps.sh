#!/usr/bin/env bash

set +x

###
# CONFIG VARIABLES
THOST='localhost'
TPORT='8080'
#CURL_FLAGS='-vvv'
CURL_FLAGS=''

###
# Default objects
CHIRP='{"body":"This is the created Chirp!"}'
USER='{
  "email": "user@example.com"
}'

function post_chirp() {
	echo "<> POST CHIRP"
	curl $CURL_FLAGS -X POST "$THOST:$TPORT/api/chirps" -d "$CHIRP"
}

function get_chirps() {
	echo "<> GET CHIRPS"
	curl $CURL_FLAGS "$THOST:$TPORT/api/chirps"
}

function post_user() {
	echo "<> POST USER"
	curl $CURL_FLAGS -X POST "$THOST:$TPORT/api/users" -d "$USER"
}

function get_users() {
	echo "<> GET USERS"
	curl $CURL_FLAGS "$THOST:$TPORT/api/users"
}

function delete_db() {
	echo "<> DELETE DB"
	curl $CURL_FLAGS -X DELETE "$THOST:$TPORT/api/db"
}

function all() {
	set -x # set -o xtrace # Same command
	echo # TODO
}

#############################################
# RUN FUNCTION WITH NAME AS FIRST PARAMETER #
#############################################
$1
