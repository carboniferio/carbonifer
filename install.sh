#!/bin/bash

REPO=carbonifer
USER=carboniferio
BINARY_NAME=carbonifer

# Defaults
DEST_DIR=${DEST_DIR:-/usr/local/bin}
VERSION=${VERSION:-$(curl -s https://api.github.com/repos/$USER/$REPO/releases/latest | grep "tag_name" | cut -d : -f 2,3 | tr -d \",[:space:])}

# Named parameters
for arg in "$@"
do
    case $arg in
        --dest-dir=*)
        DEST_DIR="${arg#*=}"
        shift
        ;;
        --version=*)
        VERSION="${arg#*=}"
        shift
        ;;
    esac
done

OS=$(uname -s)
ARCH=$(uname -m)

# Convert OS and ARCH as needed for file naming
if [ "$OS" == "Darwin" ]; then
  OS="darwin"
elif [ "$OS" == "Linux" ]; then
  OS="linux"
fi

if [ "$ARCH" == "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" == "arm64" ]; then
  ARCH="arm64"
elif [ "$ARCH" == "aarch64" ]; then
  ARCH="arm64"
fi

# Convert OS name to Title Case
OS_TITLE=$(echo "$OS" | awk '{print toupper(substr($0,1,1))tolower(substr($0,2))}')

# Construct the file name
FILE_NAME="${BINARY_NAME}_${OS_TITLE}_${ARCH}.tar.gz"

# Construct download URL
URL="https://github.com/$USER/$REPO/releases/download/$VERSION/$FILE_NAME"

echo $URL

TEMP_DIR=$(mktemp -d)
curl -sL $URL | tar xz -C $TEMP_DIR
mv $TEMP_DIR/$BINARY_NAME $DEST_DIR
rm -rf $TEMP_DIR
