module opendev.org/airship/airshipctl

go 1.13

require (
	github.com/Microsoft/go-winio v0.4.12 // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0
	github.com/docker/go-connections v0.3.0 // indirect
	github.com/docker/spdystream v0.0.0-20181023171402-6480d4af844c // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/ghodss/yaml v1.0.1-0.20190212211648-25d852aebe32 // indirect
	github.com/go-git/go-billy/v5 v5.0.0
	github.com/go-git/go-git-fixtures/v4 v4.0.1
	github.com/go-git/go-git/v5 v5.0.0
	github.com/gorilla/mux v1.7.4 // indirect
	github.com/gregjones/httpcache v0.0.0-20190212212710-3befbb6ad0cc // indirect
	github.com/morikuni/aec v0.0.0-20170113033406-39771216ff4c // indirect
	github.com/opencontainers/image-spec v1.0.1 // indirect
	github.com/spf13/afero v1.2.2
	github.com/spf13/cobra v0.0.6
	github.com/spf13/viper v1.6.2
	github.com/stretchr/testify v1.4.0
	k8s.io/api v0.17.3
	k8s.io/apiextensions-apiserver v0.17.3
	k8s.io/apimachinery v0.17.3
	k8s.io/cli-runtime v0.17.3
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/kubectl v0.17.3
	opendev.org/airship/go-redfish v0.0.0-20200318103738-db034d1d753a
	opendev.org/airship/go-redfish/client v0.0.0-20200318103738-db034d1d753a
	sigs.k8s.io/cluster-api v0.3.3
	sigs.k8s.io/kustomize/api v0.3.1
	sigs.k8s.io/yaml v1.2.0
)

replace (
	github.com/russross/blackfriday => github.com/russross/blackfriday v1.5.2
	k8s.io/client-go => k8s.io/client-go v0.0.0-20191114101535-6c5935290e33
)
