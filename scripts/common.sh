#!/usr/bin/env sh

if [ -n "$1" ] && [ ${0:0:4} = "/bin" ]; then
  ROOT_DIR=$1/..
else
  ROOT_DIR="$( cd "$( dirname "$0" )" && pwd )/.."
fi


GO_IMAGE=p1hub/go
GO_IMAGE_TAG=1.12
DIND_IMAGE=p1hub/dind
DIND_IMAGE_TAG=latest
GO_PKG=github.com/paysuper/paysuper-management-api
GOOS="linux"
GOARCH="amd64"
PROJECT_NAME="paysuper-management-api"
DOCKER_NETWORK="p1devnet"
DOCKER_IMAGE=p1hub/p1payapi