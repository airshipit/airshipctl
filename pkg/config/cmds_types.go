package config

type ClusterOptions struct {
	Name                  string
	ClusterType           string
	Server                string
	InsecureSkipTLSVerify bool
	CertificateAuthority  string
	EmbedCAData           bool
}
