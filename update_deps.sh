#!/bin/bash
# Update all the dependencies at once
find . -name go.mod -print -execdir sh -c 'pwd; go get -u ./... && go mod tidy' \;
