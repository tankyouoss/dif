#!/usr/bin/env bash
set -eo pipefail

########################################################################################################################
#
# Vars that should not be edited
#
########################################################################################################################

RED=$(printf "\33[31m")
GREEN=$(printf "\33[32m")
WHITE=$(printf "\33[37m")
YELLOW=$(printf "\33[33m")
RESET=$(printf "\33[0m")

# shellcheck disable=SC2120
function build() {
  rm -R build || true
	GOOS=darwin GOARCH=amd64 go build -o build/dif-${VERSION-latest}-osx64/dif
	GOOS=linux GOARCH=amd64 go build -o build/dif-${VERSION-latest}-linux64/dif
}

function package() {
  cd build/dif-${VERSION-latest}-osx64 && zip -r ../dif-${VERSION-latest}-osx64.zip dif && cd ../../
	cd build/dif-${VERSION-latest}-linux64 && zip -r ../dif-${VERSION-latest}-linux64.zip dif && cd ../../
}

function run() {
  go run main.go $*
}

function help {
    echo "$0 <task> <args>"
    compgen -A function | cat -n
}

#TIMEFORMAT="Task completed in %3lR"
#time ${@:-help}
"${@:-help}"