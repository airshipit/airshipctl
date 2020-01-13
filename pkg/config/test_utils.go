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

package config

import (
	"io/ioutil"
	"net/url"
	"path/filepath"
	"testing"

	"k8s.io/client-go/tools/clientcmd"
	kubeconfig "k8s.io/client-go/tools/clientcmd/api"

	"github.com/stretchr/testify/require"
)

// DummyConfig used by tests, to initialize min set of data
func DummyConfig() *Config {
	conf := &Config{
		Kind:       AirshipConfigKind,
		APIVersion: AirshipConfigApiVersion,
		Clusters: map[string]*ClusterPurpose{
			"dummy_cluster": DummyClusterPurpose(),
		},
		AuthInfos: map[string]*AuthInfo{
			"dummy_user": DummyAuthInfo(),
		},
		Contexts: map[string]*Context{
			"dummy_context": DummyContext(),
		},
		Manifests: map[string]*Manifest{
			"dummy_manifest": DummyManifest(),
		},
		ModulesConfig:  DummyModules(),
		CurrentContext: "dummy_context",
		kubeConfig:     kubeconfig.NewConfig(),
	}
	dummyCluster := conf.Clusters["dummy_cluster"]
	conf.KubeConfig().Clusters["dummycluster_target"] = dummyCluster.ClusterTypes[Target].KubeCluster()
	conf.KubeConfig().Clusters["dummycluster_ephemeral"] = dummyCluster.ClusterTypes[Ephemeral].KubeCluster()
	return conf
}

// DummyContext , utility function used for tests
func DummyContext() *Context {
	c := NewContext()
	c.NameInKubeconf = "dummy_cluster"
	c.Manifest = "dummy_manifest"
	context := kubeconfig.NewContext()
	context.Namespace = "dummy_namespace"
	context.AuthInfo = "dummy_user"
	context.Cluster = "dummycluster_ephemeral"
	c.SetKubeContext(context)

	return c
}

// DummyCluster, utility function used for tests
func DummyCluster() *Cluster {
	c := NewCluster()

	cluster := kubeconfig.NewCluster()
	cluster.Server = "http://dummy.server"
	cluster.InsecureSkipTLSVerify = false
	cluster.CertificateAuthority = "dummy_ca"
	c.SetKubeCluster(cluster)
	c.NameInKubeconf = "dummycluster_target"
	c.Bootstrap = "dummy_bootstrap"
	return c
}

// DummyManifest , utility function used for tests
func DummyManifest() *Manifest {
	m := NewManifest()
	// Repositories is the map of repository adddressable by a name
	m.Repositories["dummy"] = DummyRepository()
	m.TargetPath = "/var/tmp/"
	return m
}

func DummyRepository() *Repository {
	// TODO(howell): handle this error
	//nolint: errcheck
	url, _ := url.Parse("http://dummy.url.com")
	return &Repository{
		Url:        url,
		Username:   "dummy_user",
		TargetPath: "dummy_targetpath",
	}
}

func DummyAuthInfo() *AuthInfo {
	return NewAuthInfo()
}

func DummyModules() *Modules {
	return &Modules{Dummy: "dummy-module"}
}

// DummyClusterPurpose , utility function used for tests
func DummyClusterPurpose() *ClusterPurpose {
	cp := NewClusterPurpose()
	cp.ClusterTypes["ephemeral"] = DummyCluster()
	cp.ClusterTypes["ephemeral"].NameInKubeconf = "dummycluster_ephemeral"
	cp.ClusterTypes["target"] = DummyCluster()
	return cp
}

func InitConfig(t *testing.T) *Config {
	t.Helper()
	testDir, err := ioutil.TempDir("", "airship-test")
	require.NoError(t, err)

	configPath := filepath.Join(testDir, "config")
	err = ioutil.WriteFile(configPath, []byte(testConfigYAML), 0666)
	require.NoError(t, err)

	kubeConfigPath := filepath.Join(testDir, "kubeconfig")
	err = ioutil.WriteFile(kubeConfigPath, []byte(testKubeConfigYAML), 0666)
	require.NoError(t, err)

	conf := NewConfig()

	kubePathOptions := clientcmd.NewDefaultPathOptions()
	kubePathOptions.GlobalFile = kubeConfigPath
	err = conf.LoadConfig(configPath, kubePathOptions)
	require.NoError(t, err)

	return conf
}

func DummyClusterOptions() *ClusterOptions {
	co := &ClusterOptions{}
	co.Name = "dummy_Cluster"
	co.ClusterType = Ephemeral
	co.Server = "http://1.1.1.1"
	co.InsecureSkipTLSVerify = false
	co.CertificateAuthority = ""
	co.EmbedCAData = false

	return co
}

func DummyContextOptions() *ContextOptions {
	co := &ContextOptions{}
	co.Name = "dummy_context"
	co.Manifest = "dummy_manifest"
	co.AuthInfo = "dummy_user"
	co.CurrentContext = false
	co.Namespace = "dummy_namespace"

	return co
}

const (
	testConfigYAML = `apiVersion: airshipit.org/v1alpha1
clusters:
  def:
    cluster-type:
      ephemeral:
        bootstrap-info: ""
        cluster-kubeconf: def_ephemeral
      target:
        bootstrap-info: ""
        cluster-kubeconf: def_target
  onlyinkubeconf:
    cluster-type:
      target:
        bootstrap-info: ""
        cluster-kubeconf: onlyinkubeconf_target
  wrongonlyinconfig:
    cluster-type: {}
  wrongonlyinkubeconf:
    cluster-type:
      target:
        bootstrap-info: ""
        cluster-kubeconf: wrongonlyinkubeconf_target
contexts:
  def_ephemeral:
    context-kubeconf: def_ephemeral
  def_target:
    context-kubeconf: def_target
  onlyink:
    context-kubeconf: onlyinkubeconf_target
current-context: ""
kind: Config
manifests: {}
modules-config:
  dummy-for-tests: ""
users:
  k-admin: {}
  k-other: {}`

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
- name: k-admin
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=
- name: k-other
  user:
    client-certificate-data: LS0tLS1CRUdJTiBDRVJUSUZJQ0FURS0tLS0tCk1JSUM4akNDQWRxZ0F3SUJBZ0lJQXhEdzk2RUY4SXN3RFFZSktvWklodmNOQVFFTEJRQXdGVEVUTUJFR0ExVUUKQXhNS2EzVmlaWEp1WlhSbGN6QWVGdzB4T1RBNU1qa3hOekF6TURsYUZ3MHlNREE1TWpneE56QXpNVEphTURReApGekFWQmdOVkJBb1REbk41YzNSbGJUcHRZWE4wWlhKek1Sa3dGd1lEVlFRREV4QnJkV0psY201bGRHVnpMV0ZrCmJXbHVNSUlCSWpBTkJna3Foa2lHOXcwQkFRRUZBQU9DQVE4QU1JSUJDZ0tDQVFFQXV6R0pZdlBaNkRvaTQyMUQKSzhXSmFaQ25OQWQycXo1cC8wNDJvRnpRUGJyQWd6RTJxWVZrek9MOHhBVmVSN1NONXdXb1RXRXlGOEVWN3JyLwo0K0hoSEdpcTVQbXF1SUZ5enpuNi9JWmM4alU5eEVmenZpa2NpckxmVTR2UlhKUXdWd2dBU05sMkFXQUloMmRECmRUcmpCQ2ZpS1dNSHlqMFJiSGFsc0J6T3BnVC9IVHYzR1F6blVRekZLdjJkajVWMU5rUy9ESGp5UlJKK0VMNlEKQlltR3NlZzVQNE5iQzllYnVpcG1NVEFxL0p1bU9vb2QrRmpMMm5acUw2Zkk2ZkJ0RjVPR2xwQ0IxWUo4ZnpDdApHUVFaN0hUSWJkYjJ0cDQzRlZPaHlRYlZjSHFUQTA0UEoxNSswV0F5bVVKVXo4WEE1NDRyL2J2NzRKY0pVUkZoCmFyWmlRd0lEQVFBQm95Y3dKVEFPQmdOVkhROEJBZjhFQkFNQ0JhQXdFd1lEVlIwbEJBd3dDZ1lJS3dZQkJRVUgKQXdJd0RRWUpLb1pJaHZjTkFRRUxCUUFEZ2dFQkFMMmhIUmVibEl2VHJTMFNmUVg1RG9ueVVhNy84aTg1endVWApSd3dqdzFuS0U0NDJKbWZWRGZ5b0hRYUM4Ti9MQkxyUXM0U0lqU1JYdmFHU1dSQnRnT1RRV21Db1laMXdSbjdwCndDTXZQTERJdHNWWm90SEZpUFl2b1lHWFFUSXA3YlROMmg1OEJaaEZ3d25nWUovT04zeG1rd29IN1IxYmVxWEYKWHF1TTluekhESk41VlZub1lQR09yRHMwWlg1RnNxNGtWVU0wVExNQm9qN1ZIRDhmU0E5RjRYNU4yMldsZnNPMAo4aksrRFJDWTAyaHBrYTZQQ0pQS0lNOEJaMUFSMG9ZakZxT0plcXpPTjBqcnpYWHh4S2pHVFVUb1BldVA5dCtCCjJOMVA1TnI4a2oxM0lrend5Q1NZclFVN09ZM3ltZmJobHkrcXZxaFVFa014MlQ1SkpmQT0KLS0tLS1FTkQgQ0VSVElGSUNBVEUtLS0tLQo=
    client-key-data: LS0tLS1CRUdJTiBSU0EgUFJJVkFURSBLRVktLS0tLQpNSUlFcFFJQkFBS0NBUUVBdXpHSll2UFo2RG9pNDIxREs4V0phWkNuTkFkMnF6NXAvMDQyb0Z6UVBickFnekUyCnFZVmt6T0w4eEFWZVI3U041d1dvVFdFeUY4RVY3cnIvNCtIaEhHaXE1UG1xdUlGeXp6bjYvSVpjOGpVOXhFZnoKdmlrY2lyTGZVNHZSWEpRd1Z3Z0FTTmwyQVdBSWgyZERkVHJqQkNmaUtXTUh5ajBSYkhhbHNCek9wZ1QvSFR2MwpHUXpuVVF6Rkt2MmRqNVYxTmtTL0RIanlSUkorRUw2UUJZbUdzZWc1UDROYkM5ZWJ1aXBtTVRBcS9KdW1Pb29kCitGakwyblpxTDZmSTZmQnRGNU9HbHBDQjFZSjhmekN0R1FRWjdIVEliZGIydHA0M0ZWT2h5UWJWY0hxVEEwNFAKSjE1KzBXQXltVUpVejhYQTU0NHIvYnY3NEpjSlVSRmhhclppUXdJREFRQUJBb0lCQVFDU0pycjlaeVpiQ2dqegpSL3VKMFZEWCt2aVF4c01BTUZyUjJsOE1GV3NBeHk1SFA4Vk4xYmc5djN0YUVGYnI1U3hsa3lVMFJRNjNQU25DCm1uM3ZqZ3dVQWlScllnTEl5MGk0UXF5VFBOU1V4cnpTNHRxTFBjM3EvSDBnM2FrNGZ2cSsrS0JBUUlqQnloamUKbnVFc1JpMjRzT3NESlM2UDE5NGlzUC9yNEpIM1M5bFZGbkVuOGxUR2c0M1kvMFZoMXl0cnkvdDljWjR5ZUNpNwpjMHFEaTZZcXJZaFZhSW9RRW1VQjdsbHRFZkZzb3l4VDR6RTE5U3pVbkRoMmxjYTF1TzhqcmI4d2xHTzBoQ2JyClB1R1l2WFFQa3Q0VlNmalhvdGJ3d2lBNFRCVERCRzU1bHp6MmNKeS9zSS8zSHlYbEMxcTdXUmRuQVhhZ1F0VzkKOE9DZGRkb0JBb0dCQU5NcUNtSW94REtyckhZZFRxT1M1ZFN4cVMxL0NUN3ZYZ0pScXBqd2Y4WHA2WHo0KzIvTAozVXFaVDBEL3dGTkZkc1Z4eFYxMnNYMUdwMHFWZVlKRld5OVlCaHVSWGpTZ0ZEWldSY1Z1Y01sNVpPTmJsbmZGCjVKQ0xnNXFMZ1g5VTNSRnJrR3A0R241UDQxamg4TnhKVlhzZG5xWE9xNTFUK1RRT1UzdkpGQjc1QW9HQkFPTHcKalp1cnZtVkZyTHdaVGgvRDNpWll5SVV0ZUljZ2NKLzlzbTh6L0pPRmRIbFd4dGRHUFVzYVd1MnBTNEhvckFtbgpqTm4vSTluUXd3enZ3MWUzVVFPbUhMRjVBczk4VU5hbk5TQ0xNMW1yaXZHRXJ1VHFnTDM1bU41eFZPdTUxQU5JCm4yNkFtODBJT2JDeEtLa0R0ZXJSaFhHd3g5c1pONVJCbG9VRThZNGJBb0dBQ3ZsdVhMZWRxcng5VkE0bDNoNXUKVDJXRVUxYjgxZ1orcmtRc1I1S0lNWEw4cllBTElUNUpHKzFuendyN3BkaEFXZmFWdVV2SDRhamdYT0h6MUs5aQpFODNSVTNGMG9ldUg0V01PY1RwU0prWm0xZUlXcWRiaEVCb1FGdUlWTXRib1BsV0d4ZUhFRHJoOEtreGp4aThSCmdEcUQyajRwY1IzQ0g5QjJ5a0lqQjVFQ2dZRUExc0xXLys2enE1c1lNSm14K1JXZThhTXJmL3pjQnVTSU1LQWgKY0dNK0wwMG9RSHdDaUU4TVNqcVN1ajV3R214YUFuanhMb3ZwSFlRV1VmUEVaUW95UE1YQ2VhRVBLOU4xbk8xMwp0V2lHRytIZkIxaU5PazFCc0lhNFNDbndOM1FRVTFzeXBaeEgxT3hueS9LYmkvYmEvWEZ5VzNqMGFUK2YvVWxrCmJGV1ZVdWtDZ1lFQTBaMmRTTFlmTjV5eFNtYk5xMWVqZXdWd1BjRzQxR2hQclNUZEJxdHFac1doWGE3aDdLTWEKeHdvamh5SXpnTXNyK2tXODdlajhDQ2h0d21sQ1p5QU92QmdOZytncnJ1cEZLM3FOSkpKeU9YREdHckdpbzZmTQp5aXB3Q2tZVGVxRThpZ1J6UkI5QkdFUGY4eVpjMUtwdmZhUDVhM0lRZmxiV0czbGpUemNNZVZjPQotLS0tLUVORCBSU0EgUFJJVkFURSBLRVktLS0tLQo=`
)
