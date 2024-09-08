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
  "email": "user@example.com",
  "password": "1234"
}'
WRONG_PASS='{
  "email": "user@example.com",
  "password": "1235"
}'
WRONG_USER='{
  "email": "user.other@example.com",
  "password": "1235"
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

function post_login() {
	echo "<> POST LOGIN"
	curl $CURL_FLAGS -X POST "$THOST:$TPORT/api/login" -d "$USER"
}

function post_login_wrong_pass() {
	echo "<> POST LOGIN: Wrong password"
	curl $CURL_FLAGS -X POST "$THOST:$TPORT/api/login" -d "$WRONG_PASS"
}

function post_login_wrong_user() {
	echo "<> POST LOGIN: Wrong password"
	curl $CURL_FLAGS -X POST "$THOST:$TPORT/api/login" -d "$WRONG_USER"
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
