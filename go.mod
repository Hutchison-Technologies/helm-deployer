module github.com/Hutchison-Technologies/helm-deployer

go 1.16

replace k8s.io/api => k8s.io/api v0.21.0

replace k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.21.0

replace k8s.io/apimachinery => k8s.io/apimachinery v0.21.1-rc.0

replace k8s.io/apiserver => k8s.io/apiserver v0.21.0

replace k8s.io/cli-runtime => k8s.io/cli-runtime v0.21.0

replace k8s.io/client-go => k8s.io/client-go v0.21.0

replace k8s.io/cloud-provider => k8s.io/cloud-provider v0.21.0

replace k8s.io/cluster-bootstrap => k8s.io/cluster-bootstrap v0.21.0

replace k8s.io/code-generator => k8s.io/code-generator v0.21.2-rc.0

replace k8s.io/component-base => k8s.io/component-base v0.21.0

replace k8s.io/component-helpers => k8s.io/component-helpers v0.21.0

replace k8s.io/controller-manager => k8s.io/controller-manager v0.21.0

replace k8s.io/cri-api => k8s.io/cri-api v0.21.2-rc.0

replace k8s.io/csi-translation-lib => k8s.io/csi-translation-lib v0.21.0

replace k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.21.0

replace k8s.io/kube-controller-manager => k8s.io/kube-controller-manager v0.21.0

replace k8s.io/kube-proxy => k8s.io/kube-proxy v0.21.0

replace k8s.io/kube-scheduler => k8s.io/kube-scheduler v0.21.0

replace k8s.io/kubectl => k8s.io/kubectl v0.21.0

replace k8s.io/kubelet => k8s.io/kubelet v0.21.0

replace k8s.io/legacy-cloud-providers => k8s.io/legacy-cloud-providers v0.21.0

replace k8s.io/metrics => k8s.io/metrics v0.21.0

replace k8s.io/mount-utils => k8s.io/mount-utils v0.21.1-rc.0

replace k8s.io/sample-apiserver => k8s.io/sample-apiserver v0.21.0

replace k8s.io/sample-cli-plugin => k8s.io/sample-cli-plugin v0.21.0

replace k8s.io/sample-controller => k8s.io/sample-controller v0.21.0

require (
	github.com/databus23/helm-diff v3.1.1+incompatible
	github.com/ghodss/yaml v1.0.0
	github.com/golang/protobuf v1.23.0
	github.com/mattn/go-isatty v0.0.9 // indirect
	github.com/mgutz/ansi v0.0.0-20200706080929-d51e80ef957d // indirect
	github.com/opencontainers/runc v1.0.0-rc10 // indirect
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v1 v1.0.0-20140924161607-9f9df34309c0
	helm.sh/helm/v3 v3.6.3
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/helm v2.17.0+incompatible
	rsc.io/letsencrypt v0.0.3 // indirect
	sigs.k8s.io/kustomize/kyaml v0.11.0 // indirect
)
