#!/bin/bash

OUTPUT_DIR=$PWD/dist
mkdir -p ${OUTPUT_DIR}

CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o ${OUTPUT_DIR}/nginx2sfx -ldflags "-X github.com/rabobank/nginx2sfx/conf.VERSION=${VERSION} -X github.com/rabobank/nginx2sfx/conf.COMMIT=${COMMIT}" .
