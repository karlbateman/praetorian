#!/bin/bash
go test -run=^$ -fuzz=. -fuzztime="${FUZZTIME:-30s}" .
