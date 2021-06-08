#/bin/env bash

golangci-lint run ./...\
 -E golint -E govet -E goimports -E maligned -E prealloc -E errcheck -E staticcheck\
 -D forbidigo -D gochecknoglobals -D wsl -D exhaustivestruct -D dogsled -D dupl\
 -p bugs -p complexity -p format -p performance -p style -p unused\
 --fix --sort-results
