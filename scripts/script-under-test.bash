#!/bin/bash

set -e
set -o pipefail

function capitalize() {
  for var in "$@"; do
    echo -n $var "" | tr '[:lower:]' '[:upper:]'
  done
}

function readInput() {
  input=$(cat)
  echo $input
}

function convertToTitle() {
  echo "this should be in title case" | make-title-case
}

function testSpies() {
  echo $PATH
  cf version
  cf push "$@"
}

function testStub() {
  false
}

function testMocks() {
  cf help
  bosh --version
  bosh envs
}
