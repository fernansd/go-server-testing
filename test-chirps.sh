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

function hola() {
	echo "hola"
}

function all() {
	set -x # set -o xtrace # Same command
	echo # TODO
}

#############################################
# RUN FUNCTION WITH NAME AS FIRST PARAMETER #
#############################################
$1
