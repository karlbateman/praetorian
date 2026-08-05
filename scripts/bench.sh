#!/usr/bin/env bash
go test -run=^$ -bench=. -benchmem .
