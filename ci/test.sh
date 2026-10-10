#!/bin/sh
set -eu
mkdir -p /reports
# -count=1 always executes tests instead of accepting Go's cached results.
result=0
go test -count=1 -json ./... > /reports/go-test.json || result=$?
cat /reports/go-test.json
go run ./ci/go-junit.go /reports/go-test.json /reports/junit.xml
exit "$result"
