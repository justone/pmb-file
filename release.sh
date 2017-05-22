#!/bin/bash

ORG=justone
NAME=pmb-file

set -e

if [[ ! $(type -P gox) ]]; then
    echo "Error: gox not found."
    echo "To fix: run 'go get github.com/mitchellh/gox', and/or add \$GOPATH/bin to \$PATH"
    exit 1
fi

if [[ ! $(type -P github-release) ]]; then
    echo "Error: github-release not found."
    exit 1
fi

VER=$1

if [[ -z $VER ]]; then
    echo "Need to specify version."
    exit 1
fi

PRE_ARG=
if [[ $VER =~ pre ]]; then
    PRE_ARG="--pre-release"
fi

git tag $VER

echo "Building $VER"
echo

gox -ldflags "-X main.version=$VER" -osarch="darwin/amd64 linux/amd64 linux/arm windows/amd64"

echo "* " > desc
echo "" >> desc

echo "$ sha1sum ${NAME}_*" >> desc
sha1sum ${NAME}_* >> desc
echo "$ sha256sum ${NAME}_*" >> desc
sha256sum ${NAME}_* >> desc
echo "$ md5sum ${NAME}_*" >> desc
md5sum ${NAME}_* >> desc

vi desc

git push --tags

sleep 2

cat desc | github-release release $PRE_ARG --user ${ORG} --repo ${NAME} --tag $VER --name $VER --description -
github-release upload --user ${ORG} --repo ${NAME} --tag $VER --name ${NAME}_darwin_amd64 --file ${NAME}_darwin_amd64
github-release upload --user ${ORG} --repo ${NAME} --tag $VER --name ${NAME}_linux_amd64 --file ${NAME}_linux_amd64
github-release upload --user ${ORG} --repo ${NAME} --tag $VER --name ${NAME}_windows_amd64.exe --file ${NAME}_windows_amd64.exe
github-release upload --user ${ORG} --repo ${NAME} --tag $VER --name ${NAME}_linux_arm --file ${NAME}_linux_arm
