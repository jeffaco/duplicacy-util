#!/usr/bin/env bash

perform_build()
{
    # Parameters to supply:
    #   $1: Operating System (GOOS)
    #   $2: Architecture (GOARCH)

    echo "Building for: $1 $2"

    # Determine output filename

    EXTENSION=

    case $1 in
        darwin)
            OSNAME=osx
            ;;
        windows)
            OSNAME=win
            EXTENSION=.exe
            ;;
        *)
            OSNAME=$1
    esac

    case $2 in
        386)
            ARCH=i386
            ;;
        amd64)
            ARCH=x64
            ;;
        *)
            ARCH=$2
    esac

    GOOS=$1 GOARCH=$2 go build \
        -ldflags "-X main.versionText=$VERSION -X main.gitHash=$GITHASH" \
        -o duplicacy-util_${OSNAME}_${ARCH}_${VERSION}${EXTENSION}
}

# Grab the version number

source version
GITHASH=`git rev-parse --short HEAD`
echo "Version: $VERSION, hash: $GITHASH"
echo

# Now build the various versions

perform_build freebsd amd64
perform_build linux arm
perform_build linux 386
perform_build linux amd64
perform_build darwin amd64
perform_build windows 386
perform_build windows amd64
