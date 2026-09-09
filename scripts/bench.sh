#!/bin/bash
go test -run=^$ -bench=. -benchmem .
