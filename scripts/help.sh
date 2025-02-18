#!/usr/bin/env bash
grep -E '^[a-zA-Z_-]+:.*?## .*$$' "$(MAKEFILE_LIST)" | \
  sed 's/Makefile://' | \
  sort -d | \
  awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'