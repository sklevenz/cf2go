#!/bin/bash

REV_FILE=".revision"
REV_GO_FILE="revision.go"

export REV="$(cat $REV_FILE)"
REV=$(expr $REV + 1)

echo $REV
echo $REV > $REV_FILE

cat > $REV_GO_FILE <<EOF 
package main

// file generated

// REVISION of software version
var REVISION="$REV"
EOF

git add $REV_FILE
git add $REV_GO_FILE
git commit -m "revision $(cat $REV_FILE)"
