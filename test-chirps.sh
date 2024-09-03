#!/usr/bin/env bash

set +x

###
# CONFIG VARIABLES
THOST='localhost'
TPORT='8080'

###
# Default objects
CHIRP='{"body":"This is the created Chirp!"}'
USER='{
  "email": "user@example.com"
}'

function get_chirps() {
	echo "<> GET CHIRPS"
	curl -vvv "$THOST:$TPORT/api/chirps"
}

function post_user() {
	echo "<> POST USER"
	curl -vvv -X POST "$THOST:$TPORT/api/users" -d "$USER"
}

function delete_db() {
	echo "<> DELETE DB"
	curl -vvv -X DELETE "$THOST:$TPORT/api/db"
}

function all() {
	set -x # set -o xtrace # Same command
	echo # TODO
}

#############################################
# RUN FUNCTION WITH NAME AS FIRST PARAMETER #
#############################################
$1
