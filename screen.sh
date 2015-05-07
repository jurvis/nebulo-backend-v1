#!/bin/bash

echo "Setting environment variables GOBIN and GOPATH..."
cd ./bin
export GOBIN=$(pwd)
cd ../
export GOPATH=$(pwd)
echo "Rebuilding..."
go install -a
echo "Running..."
screen -S nebulo-backend ./bin/nebulo-backend