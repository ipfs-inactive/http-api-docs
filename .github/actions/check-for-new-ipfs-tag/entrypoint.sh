#!/usr/bin/env sh
set -eu

# extract go-ipfs release tag used in http-api-docs from go.mod in this repo
CURRENT_IPFS_TAG = $(grep 'github.com/ipfs/go-ipfs ' ./http-api-docs/go.mod | awk '{print $2}')

echo "The currently used IPFS tag is ${CURRENT_IPFS_TAG}"

# XXX: extract release XXX
