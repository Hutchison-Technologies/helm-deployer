#!/bin/bash

VERSION=${1#"v"}
if [ -z "$VERSION" ]; then
  echo "Please specify the Kubernetes version: e.g."
  echo "./download-deps.sh v1.21.0"
  exit 1
fi

set -euo pipefail

# Find out all the replaced imports, make a list of them.
MODS=($(
  curl -sS "https://raw.githubusercontent.com/kubernetes/kubernetes/v${VERSION}/go.mod" |
    sed -n 's|.*k8s.io/\(.*\) => ./staging/src/k8s.io/.*|k8s.io/\1|p'
))

# Now add those similar replace statements in the local go.mod file, but first find the version that
# the Kubernetes is using for them.
for MOD in "${MODS[@]}"; do
  V=$(
    go mod download -json "${MOD}@kubernetes-${VERSION}" |
      sed -n 's|.*"Version": "\(.*\)".*|\1|p'
  )

  go mod edit "-replace=${MOD}=${MOD}@${V}"
done

go get "k8s.io/kubernetes@v${VERSION}"

# Helm dependencies
go get helm.sh/helm/v3/pkg/action
go get helm.sh/helm/v3/pkg/chart/loader
go get helm.sh/helm/v3/pkg/cli
go get helm.sh/helm/v3/pkg/kube

# More k8s dependencies
go get k8s.io/api/core/v1@v0.21.0
go get k8s.io/apimachinery/pkg/apis/meta/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/apps/v1@v0.21.0
# go get k8s.io/client-go/kubernetes/typed/autoscaling/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/core/v1@v0.21.0
go get k8s.io/client-go/plugin/pkg/client/auth/gcp@v0.21.0
go get k8s.io/client-go/util/retry@v0.21.0
go get k8s.io/api/core/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/core/v1@v0.21.0
go get k8s.io/api/core/v1@v0.21.0
go get k8s.io/apimachinery/pkg/apis/meta/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/core/v1@v0.21.0
go get k8s.io/client-go/kubernetes@v0.21.0
go get k8s.io/client-go/kubernetes/typed/apps/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/autoscaling/v1@v0.21.0
go get k8s.io/client-go/kubernetes/typed/core/v1@v0.21.0
go get k8s.io/client-go/plugin/pkg/client/auth/gcp@v0.21.0
go get k8s.io/client-go/rest@v0.21.0
go get k8s.io/client-go/tools/clientcmd@v0.21.0

# even more helm and k8s dependencies
go get github.com/databus23/helm-diff/diff
go get github.com/databus23/helm-diff/manifest
go get gopkg.in/yaml.v1
go get k8s.io/helm/pkg/storage/errors
go get github.com/ghodss/yaml

go mod download