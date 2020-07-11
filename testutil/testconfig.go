/*
Copyright 2014 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package testutil

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/require"

	"opendev.org/airship/airshipctl/pkg/config"
	"opendev.org/airship/airshipctl/pkg/remote/redfish"
)

// types cloned directory from pkg/config/types to prevent circular import

// DummyConfig used by tests, to initialize min set of data
func DummyConfig() *config.Config {
	conf := &config.Config{
		Kind:       config.AirshipConfigKind,
		APIVersion: config.AirshipConfigAPIVersion,
		Clusters: map[string]*config.ClusterPurpose{
			"dummy_cluster": DummyClusterPurpose(),
		},
		AuthInfos: map[string]*config.AuthInfo{
			"dummy_user": DummyAuthInfo(),
		},
		BootstrapInfo: map[string]*config.Bootstrap{
			"dummy_bootstrap_config": DummyBootstrapInfo(),
		},
		Contexts: map[string]*config.Context{
			"dummy_context": DummyContext(),
		},
		Manifests: map[string]*config.Manifest{
			"dummy_manifest": DummyManifest(),
		},
		ManagementConfiguration: map[string]*config.ManagementConfiguration{
			"dummy_management_config": DummyManagementConfiguration(),
		},
		CurrentContext: "dummy_context",
	}
	conf.SetKubeConfig(kubeconfig.NewConfig())

	dummyCluster := conf.Clusters["dummy_cluster"]
	conf.KubeConfig().Clusters["dummy_cluster_target"] = dummyCluster.ClusterTypes[config.Target].KubeCluster()
	conf.KubeConfig().Clusters["dummy_cluster_ephemeral"] = dummyCluster.ClusterTypes[config.Ephemeral].KubeCluster()
	return conf
}

// DummyContext creates a Context config object for unit testing
func DummyContext() *config.Context {
	c := config.NewContext()
	c.NameInKubeconf = "dummy_cluster_ephemeral"
	c.Manifest = "dummy_manifest"
	context := kubeconfig.NewContext()
	context.Namespace = "dummy_namespace"
	context.AuthInfo = "dummy_user"
	context.Cluster = "dummy_cluster_ephemeral"
	c.SetKubeContext(context)

	return c
}

// DummyCluster creates a Cluster config object for unit testing
func DummyCluster() *config.Cluster {
	c := config.NewCluster()

	cluster := kubeconfig.NewCluster()
	cluster.Server = "http://dummy.server"
	cluster.InsecureSkipTLSVerify = false
	cluster.CertificateAuthority = "dummy_ca"
	c.SetKubeCluster(cluster)
	c.NameInKubeconf = "dummy_cluster_target"
	c.Bootstrap = "dummy_bootstrap_config"
	c.ManagementConfiguration = "dummy_management_config"
	return c
}

// DummyManifest creates a Manifest config object for unit testing
func DummyManifest() *config.Manifest {
	m := config.NewManifest()
	// Repositories is the map of repository adddressable by a name
	m.Repositories = map[string]*config.Repository{"primary": DummyRepository()}
	m.PrimaryRepositoryName = "primary"
	m.TargetPath = "/var/tmp/"
	m.SubPath = "manifests/site/test-site"
	return m
}

// DummyRepository creates a Repository config object for unit testing
func DummyRepository() *config.Repository {
	return &config.Repository{
		URLString: "http://dummy.url.com/manifests.git",
		CheckoutOptions: &config.RepoCheckout{
			Tag:           "v1.0.1",
			ForceCheckout: false,
		},
		Auth: &config.RepoAuth{
			Type:    "ssh-key",
			KeyPath: "testdata/test-key.pem",
		},
	}
}

// DummyRepoAuth creates a RepoAuth config object for unit testing
func DummyRepoAuth() *config.RepoAuth {
	return &config.RepoAuth{
		Type:    "ssh-key",
		KeyPath: "testdata/test-key.pem",
	}
}

// DummyRepoCheckout creates a RepoCheckout config object
// for unit testing
func DummyRepoCheckout() *config.RepoCheckout {
	return &config.RepoCheckout{
		Tag:           "v1.0.1",
		ForceCheckout: false,
	}
}

// DummyAuthInfo creates a AuthInfo config object for unit testing
func DummyAuthInfo() *config.AuthInfo {
	a := config.NewAuthInfo()
	authinfo := kubeconfig.NewAuthInfo()
	authinfo.Username = "dummy_username"
	authinfo.Password = "dummy_password"
	authinfo.ClientCertificate = "dummy_certificate"
	authinfo.ClientKey = "dummy_key"
	authinfo.Token = "dummy_token"
	encodedAuthInfo := config.EncodeAuthInfo(authinfo)
	a.SetKubeAuthInfo(encodedAuthInfo)
	return a
}

// DummyKubeAuthInfo creates a AuthInfo kubeconfig object
// for unit testing
func DummyKubeAuthInfo() *kubeconfig.AuthInfo {
	authinfo := kubeconfig.NewAuthInfo()
	authinfo.Username = "dummy_username"
	authinfo.Password = "dummy_password"
	authinfo.ClientCertificate = "dummy_certificate"
	authinfo.ClientKey = "dummy_key"
	authinfo.Token = "dummy_token"
	return authinfo
}

// DummyClusterPurpose creates ClusterPurpose config object for unit testing
func DummyClusterPurpose() *config.ClusterPurpose {
	cp := config.NewClusterPurpose()
	cp.ClusterTypes["ephemeral"] = DummyCluster()
	cp.ClusterTypes["ephemeral"].NameInKubeconf = "dummy_cluster_ephemeral"
	cp.ClusterTypes["target"] = DummyCluster()
	return cp
}

// InitConfig creates a Config object meant for testing.
//
// The returned config object will be associated with real files stored in a
// directory in the user's temporary file storage
// This directory can be cleaned up by calling the returned "cleanup" function
func InitConfig(t *testing.T) (conf *config.Config, cleanup func(*testing.T)) {
	t.Helper()
	testDir, cleanup := TempDir(t, "airship-test")

	configPath := filepath.Join(testDir, "config")
	err := ioutil.WriteFile(configPath, []byte(testConfigYAML), 0666)
	require.NoError(t, err)

	kubeConfigPath := filepath.Join(testDir, "kubeconfig")
	err = ioutil.WriteFile(kubeConfigPath, []byte(testKubeConfigYAML), 0666)
	require.NoError(t, err)

	conf = config.NewConfig()

	err = conf.LoadConfig(configPath, kubeConfigPath, false)
	require.NoError(t, err)

	return conf, cleanup
}

// DummyClusterOptions creates ClusterOptions config object
// for unit testing
func DummyClusterOptions() *config.ClusterOptions {
	co := &config.ClusterOptions{}
	co.Name = "dummy_cluster"
	co.ClusterType = config.Ephemeral
	co.Server = "http://1.1.1.1"
	co.InsecureSkipTLSVerify = false
	co.CertificateAuthority = ""
	co.EmbedCAData = false

	return co
}

// DummyContextOptions creates ContextOptions config object
// for unit testing
func DummyContextOptions() *config.ContextOptions {
	co := &config.ContextOptions{}
	co.Name = "dummy_context"
	co.Manifest = "dummy_manifest"
	co.AuthInfo = "dummy_user"
	co.CurrentContext = false
	co.Namespace = "dummy_namespace"

	return co
}

// DummyAuthInfoOptions creates AuthInfoOptions config object
// for unit testing
func DummyAuthInfoOptions() *config.AuthInfoOptions {
	authinfo := &config.AuthInfoOptions{}
	authinfo.Username = "dummy_username"
	authinfo.Password = "dummy_password"
	authinfo.ClientCertificate = "dummy_certificate"
	authinfo.ClientKey = "dummy_key"
	authinfo.Token = "dummy_token"
	return authinfo
}

// DummyBootstrapInfo creates a dummy BootstrapInfo config object for unit testing
func DummyBootstrapInfo() *config.Bootstrap {
	bs := &config.Bootstrap{}
	cont := config.Container{
		Volume:           "/dummy:dummy",
		Image:            "dummy_image:dummy_tag",
		ContainerRuntime: "docker",
	}
	builder := config.Builder{
		UserDataFileName:       "user-data",
		NetworkConfigFileName:  "netconfig",
		OutputMetadataFileName: "output-metadata.yaml",
	}

	bs.Container = &cont
	bs.Builder = &builder

	return bs
}

// DummyManagementConfiguration creates a management configuration for unit testing
func DummyManagementConfiguration() *config.ManagementConfiguration {
	return &config.ManagementConfiguration{
		Type:     redfish.ClientType,
		Insecure: true,
		UseProxy: false,
	}
}

// DummyManifestOptions creates ManifestOptions config object
// for unit testing
func DummyManifestOptions() *config.ManifestOptions {
	return &config.ManifestOptions{
		Name:       "dummy_manifest",
		SubPath:    "manifests/dummy_site",
		TargetPath: "/tmp/dummy_site",
		IsPrimary:  true,
		RepoName:   "dummy_repo",
		URL:        "https://github.com/treasuremap/dummy_site",
		Branch:     "master",
		Force:      true,
	}
}

const (
	testConfigYAML = `apiVersion: airshipit.org/v1alpha1
bootstrapInfo:
  default: {}
clusters:
  straggler:
    clusterType:
      ephemeral:
        clusterKubeconf: notThere
  def:
    clusterType:
      ephemeral:
        bootstrapInfo: ""
        clusterKubeconf: def_ephemeral
      target:
        bootstrapInfo: ""
        clusterKubeconf: def_target
  onlyinkubeconf:
    clusterType:
      target:
        bootstrapInfo: ""
        clusterKubeconf: onlyinkubeconf_target
  wrongonlyinconfig:
    clusterType: {}
  wrongonlyinkubeconf:
    clusterType:
      target:
        bootstrapInfo: ""
        clusterKubeconf: wrongonlyinkubeconf_target
  clustertypenil:
    clusterType: null
contexts:
  def_ephemeral:
    contextKubeconf: def_ephemeral
  def_target:
    contextKubeconf: def_target
  onlyink:
    contextKubeconf: onlyinkubeconf_target
currentContext: ""
kind: Config
manifests: {}
users:
  k-admin: {}
  k-other: {}
  def-user: {}`

	//nolint:lll
	testKubeConfigYAML = `apiVersion: v1
clusters:
- cluster:
    insecure-skip-tls-verify: true
    server: http://5.6.7.8
  name: def_ephemeral
- cluster:
    insecure-skip-tls-verify: true
    server: http://1.2.3.4
  name: def_target
- cluster:
    insecure-skip-tls-verify: true
    server: http://9.10.11.12
  name: onlyinkubeconf_target
- cluster:
    certificate-authority: cert_file
    server: ""
  name: wrongonlyinkubeconf_target
- cluster:
    insecure-skip-tls-verify: true
    server: http://9.10.11.12
  name: invalidName
- cluster:
    insecure-skip-tls-verify: true
    server: http://9.10.11.12
  name: clustertypenil_target
contexts:
- context:
    cluster: def_ephemeral
    user: k-admin
  name: def_ephemeral
- context:
    cluster: def_target
    user: k-admin
  name: def_target
- context:
    cluster: onlyinkubeconf_target
    user: k-other
  name: onlyink
current-context: ""
kind: Config
preferences: {}
users:
users:
- name: def-user
  user:
     username: dummy_username
     password: ZHVtbXlfcGFzc3dvcmQK
- name: k-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
- name: k-other
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
)
