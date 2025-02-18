#!/usr/bin/env bash
go build -ldflags="-w -s" -o dist/praetorian ./cmd
go test -c -o dist/praetorian.test