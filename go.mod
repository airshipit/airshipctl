module opendev.org/airship/airshipctl

go 1.13

require (
	github.com/Azure/go-autorest/autorest v0.11.7 // indirect
	github.com/Masterminds/goutils v1.1.0 // indirect
	github.com/Masterminds/semver v1.5.0 // indirect
	github.com/Masterminds/sprig v2.22.0+incompatible
	github.com/Microsoft/go-winio v0.4.14 // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/cheggaaa/pb/v3 v3.0.4
	github.com/containerd/containerd v1.4.1 // indirect
	github.com/docker/docker v1.4.2-0.20200203170920-46ec8731fbce
	github.com/docker/go-connections v0.4.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git-fixtures/v4 v4.0.1
	github.com/go-git/go-git/v5 v5.0.0
	github.com/go-logr/zapr v0.1.1 // indirect
	github.com/gophercloud/gophercloud v0.6.0 // indirect
	github.com/gopherjs/gopherjs v0.0.0-20191106031601-ce3c9ade29de // indirect
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/hashicorp/go-cleanhttp v0.5.1 // indirect
	github.com/huandu/xstrings v1.3.1 // indirect
	github.com/mitchellh/copystructure v1.0.0 // indirect
	github.com/morikuni/aec v1.0.0 // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.0.0
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.9
	k8s.io/apiextensions-apiserver v0.17.9
	k8s.io/apimachinery v0.17.9
	k8s.io/cli-runtime v0.17.4
	k8s.io/client-go v0.17.9
	k8s.io/kubectl v0.17.4
	opendev.org/airship/go-redfish v0.0.0-20200318103738-db034d1d753a
	opendev.org/airship/go-redfish/client v0.0.0-20200318103738-db034d1d753a
	sigs.k8s.io/cli-utils v0.18.1
	sigs.k8s.io/cluster-api v0.3.10
	sigs.k8s.io/controller-runtime v0.5.11
	sigs.k8s.io/kustomize/api v0.5.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	k8s.io/kubectl => k8s.io/kubectl v0.0.0-20191219154910-1528d4eea6dd
	sigs.k8s.io/kustomize/kyaml => sigs.k8s.io/kustomize/kyaml v0.4.1
)
