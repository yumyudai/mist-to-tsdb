#!/bin/bash
cat client.json| jq '.data' | sed 's/^"//' | sed 's/"$//' | sed 's/\\"/"/g' | jq
