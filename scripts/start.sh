#!/usr/bin/env bash
DEFAULT_CONFIG=$(printf '{"activeKeyId": "1", "rootKeys": {"1": "%s"}}' "$(openssl rand -base64 32)")
PRAETORIAN_CONFIG="${PRAETORIAN_CONFIG:-$DEFAULT_CONFIG}"
export PRAETORIAN_CONFIG
go run cmd/main.go