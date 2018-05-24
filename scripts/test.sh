#!/bin/bash
set +x

. ./scripts/template.sh $1

CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .

sam local invoke CICleanerFunction -e event.json
