#!/bin/bash

#Defaults
PUBLISH="--skip-publish"
PROMOTE="--next-patch"

while getopts hgpmi option
do
 case "${option}" in
  g) PUBLISH="";;
  p) PROMOTE="--next-patch";;
  m) PROMOTE="--next-major";;
  i) PROMOTE="--next-minor";;
 esac
done

# next patch: 1.4.2 -> 1.4.3
VERSION=`git semver ${PROMOTE}`
git tag -am $VERSION $VERSION
git push 
# next minor: 1.4.2 -> 1.5.0
# VERSION=`git semver --next-minor`

# next major: 1.4.2 -> 2.0.0
# VERSION=`git semver --next-major`

goreleaser --rm-dist ${PUBLISH}
# goreleaser --rm-dist 