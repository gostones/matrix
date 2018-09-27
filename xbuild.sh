#!/usr/bin/env bash

#cross build binary
function xbuild() {
    export GOOS=linux
    export GOARCH=amd64
    export CGO_ENABLED=0

    echo "#### Cross building for $GOOS $GOARCH ..."

    build; if [ $? -ne 0 ]; then
        echo "#### Build failure"
        return 1
    fi
    return 0
}

function build() {
    echo "## Cleaning ..."
    go clean $FLAG ./...

    echo "## Vetting ..."
    go vet $FLAG ./...; if [ $? -ne 0 ]; then
        return 1
    fi

    echo "## Building ..."
    go build $FLAG -buildmode=exe -o bin/matrix -ldflags '-extldflags "-static"'; if [ $? -ne 0 ]; then
        return 1
    fi

#    echo "## Testing ..."
#    go test $FLAG ./...; if [ $? -ne 0 ]; then
#        return 1
#    fi
}
#
echo "#### Building ..."

xbuild; if [ $? -ne 0 ]; then
    echo "#### Build failure"
    exit 1
fi

echo "#### Build success"

exit 0