#!/bin/bash

set -x

GOPATH=$(go env GOPATH)
PACKAGE_NAME=k8s.io/sample-controller
REPO_ROOT="$GOPATH/src/$PACKAGE_NAME"
DOCKER_REPO_ROOT="/go/src/$PACKAGE_NAME"
DOCKER_CODEGEN_PKG="/go/src/k8s.io/code-generator"
apiGroups=(migrate/v1)

pushd $REPO_ROOT

## Generate ugorji stuff
rm "$REPO_ROOT"/pkg/apis/migrate/v1/*.deepcopy.go



# for both CRD and EAS types
docker run --rm -ti -u $(id -u):$(id -g) \
  -v "$REPO_ROOT":"$DOCKER_REPO_ROOT" \
  -w "$DOCKER_REPO_ROOT" \
  hub.tencentyun.com/xkcp0324/gengo:release-1.12 "$DOCKER_CODEGEN_PKG"/generate-groups.sh all \
  k8s.io/sample-controller/pkg/client \
  k8s.io/sample-controller/pkg/apis \
  "migrate:v1" \
  --go-header-file "$DOCKER_REPO_ROOT/hack/boilerplate.go.txt"

popd