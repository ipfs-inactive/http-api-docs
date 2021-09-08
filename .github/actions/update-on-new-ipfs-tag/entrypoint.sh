#!/usr/bin/env sh
set -eu

# extract go-ipfs release tag used in http-api-docs from go.mod in this repo
CURRENT_IPFS_TAG=`grep 'github.com/ipfs/go-ipfs ' ./go.mod | awk '{print $2}'`
echo "The currently used IPFS tag is ${CURRENT_IPFS_TAG}"

# extract IPFS release
LATEST_IPFS_TAG=$1
echo "The latest IPFS tag is ${LATEST_IPFS_TAG}"

# make the upgrade, if newer go-ipfs tags exist
if [ "$CURRENT_IPFS_TAG" = "$LATEST_IPFS_TAG" ]; then
    echo "http-api-docs already uses the latest go-ipfs tag."
else
     git checkout -b update-ipfs-to-$LATEST_IPFS_TAG
     sed "s/^\s*github.com\/ipfs\/go-ipfs\s\+$CURRENT_IPFS_TAG\s*$/	github.com\/ipfs\/go-ipfs $LATEST_IPFS_TAG/" go.mod > go.mod2
     mv go.mod2 go.mod
     go mod tidy
     make
     git add -u
     git commit -m "Bumped go-ipfs dependence to tag $LATEST_IPFS_TAG."
fi
