package config

type ClusterOptions struct {
	Name                  string
	ClusterType           string
	Server                string
	InsecureSkipTLSVerify bool
	CertificateAuthority  string
	EmbedCAData           bool
}

type ContextOptions struct {
	Name           string
	ClusterType    string
	CurrentContext bool
	Cluster        string
	AuthInfo       string
	Manifest       string
	Namespace      string
}
