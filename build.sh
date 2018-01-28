#!/bin/bash

set -eu

rm -rf Connect4.app
mkdir -p Connect4.app/Contents/MacOS
go build -o Connect4.app/Contents/MacOS/Connect4 ./cmd/connect4/*.go
