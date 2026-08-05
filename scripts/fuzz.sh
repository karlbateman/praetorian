#!/usr/bin/env bash
go test -run=^$ -fuzz=. -fuzztime="${FUZZTIME:-30s}" .
