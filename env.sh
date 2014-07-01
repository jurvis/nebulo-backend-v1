#!/usr/bin/env bash

	gp=$(pwd)/src
	if [[ -d "$gp" ]]; then
    echo "src/ exists, setting GOPATH..."
    export GOPATH=$(pwd)
  else
    export GOPATH=$(pwd)
    go get .
  fi