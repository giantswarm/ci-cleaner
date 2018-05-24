#!/bin/bash
set +x

CGO_ENABLED=0 GOOS=linux go build -a -ldflags '-extldflags "-static"' .

sam local invoke CICleanerFunction -e event.json
