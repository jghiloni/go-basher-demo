#!/bin/bash

set -e
set -o pipefail

function capitalize() {
  for var in "$@"; do
    echo -n $var "" | tr '[:lower:]' '[:upper:]'
  done
}

[[ "${BASH_SOURCE[0]}" == "${0}" ]] && capitalize "$@"
