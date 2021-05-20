/*
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

// Package v1alpha1 generates the custom resource definition schema for NetworkCatalogues
// Ignore lint for the entire file is added because there is a long regex to support IPV4 and IPV6 format.
// This regex cannot be broken down. When nolint is added for the specific line it gets picked as description
// for that field by kubebuilder controller-gen.
//nolint:lll
package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// HostNetworkingSpec defines the properties for host neworking like Links, Networks and Services
type HostNetworkingSpec struct {
	Links    []Link    `json:"links,omitempty"`
	Networks []Network `json:"networks,omitempty"`
	Services []Service `json:"services,omitempty"`
}

// Link defines the properties of the network link
type Link struct {
	ID                 string   `json:"id,omitempty"`
	Name               string   `json:"name,omitempty"`
	Type               string   `json:"type,omitempty"`
	MTU                string   `json:"mtu,omitempty"`
	BondLinks          []string `json:"bond_links,omitempty"`
	BondMode           string   `json:"bond_mode,omitempty"`
	BondMiimon         int      `json:"bond_miimon,omitempty"`
	BondXmitHashPolicy string   `json:"bond_xmit_hash_policy,omitempty"`
	VlanLink           string   `json:"vlan_link,omitempty"`
	VlanID             int      `json:"vlan_id,omitempty"`
	VlanMacAddress     string   `json:"vlan_mac_address,omitempty"`
}

// IPFormat Regex to support both IPV4 and IPV6 format
// +kubebuilder:validation:Pattern="((^((([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5])\\.){3}([0-9]|[1-9][0-9]|1[0-9]{2}|2[0-4][0-9]|25[0-5]))$)|(^(([0-9a-fA-F]{1,4}:){7,7}[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,7}:|([0-9a-fA-F]{1,4}:){1,6}:[0-9a-fA-F]{1,4}|([0-9a-fA-F]{1,4}:){1,5}(:[0-9a-fA-F]{1,4}){1,2}|([0-9a-fA-F]{1,4}:){1,4}(:[0-9a-fA-F]{1,4}){1,3}|([0-9a-fA-F]{1,4}:){1,3}(:[0-9a-fA-F]{1,4}){1,4}|([0-9a-fA-F]{1,4}:){1,2}(:[0-9a-fA-F]{1,4}){1,5}|[0-9a-fA-F]{1,4}:((:[0-9a-fA-F]{1,4}){1,6})|:((:[0-9a-fA-F]{1,4}){1,7}|:))$))"
type IPFormat string

// Network defines the network attributes like ID, Type, Link, Netmask and Routes
type Network struct {
	ID   string `json:"id,omitempty"`
	Type string `json:"type,omitempty"`
	Link string `json:"link,omitempty"`

	NetMask IPFormat `json:"netmask,omitempty"`
	Routes  []Route  `json:"routes,omitempty"`
}

// Route defines the spec for network route
type Route struct {
	Network IPFormat `json:"network,omitempty"`
	NetMask IPFormat `json:"netmask,omitempty"`
	Gateway IPFormat `json:"gateway,omitempty"`
}

// Service defines the spec for service
type Service struct {
	Address IPFormat `json:"address,omitempty"`
	Type    string   `json:"type,omitempty"`
}

// EndPointSpec defines the properties of end points like IP and port
type EndPointSpec struct {
	Host IPFormat `json:"host,omitempty"`

	// +kubebuilder:validation:Maximum=65535
	Port int `json:"port,omitempty"`
}

// KubernetesSpec defines the spec for kubernetes
type KubernetesSpec struct {
	// +kubebuilder:validation:Format=cidr
	ServiceCidr string `json:"serviceCidr,omitempty"`

	// +kubebuilder:validation:Format=cidr
	PodCidr              string       `json:"podCidr,omitempty"`
	ControlPlaneEndpoint EndPointSpec `json:"controlPlaneEndpoint,omitempty"`
	ApiserverCertSANs    string       `json:"apiserverCertSANs,omitempty"`
}

// IronicSpec defines the spec for Ironic
type IronicSpec struct {
	ProvisioningInterface   string   `json:"provisioningInterface,omitempty"`
	ProvisioningIP          IPFormat `json:"provisioningIp,omitempty"`
	DhcpRange               string   `json:"dhcpRange,omitempty"`
	IronicAutomatedClean    string   `json:"ironicAutomatedClean,omitempty"`
	HTTPPort                string   `json:"httpPort,omitempty"`
	IronicFastTrack         string   `json:"ironicFastTrack,omitempty"`
	DeployKernelURL         string   `json:"deployKernelUrl,omitempty"`
	DeployRamdiskURL        string   `json:"deployRamdiskUrl,omitempty"`
	IronicEndpoint          string   `json:"ironicEndpoint,omitempty"`
	IronicInspectorEndpoint string   `json:"ironicInspectorEndpoint,omitempty"`
}

// NtpSpec defines the spec for NTP servers
type NtpSpec struct {
	Enabled bool     `json:"enabled,omitempty"`
	Servers []string `json:"servers,omitempty"`
}

// NetworkCatalogueSpec defines the default networking catalogs hosted in airshipctl
type NetworkCatalogueSpec struct {
	CommonHostNetworking HostNetworkingSpec `json:"commonHostNetworking,omitempty"`
	Kubernetes           KubernetesSpec     `json:"kubernetes,omitempty"`
	Ironic               IronicSpec         `json:"ironic,omitempty"`
	Ntp                  NtpSpec            `json:"ntp,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkCatalogue is the Schema for the network catalogs API
type NetworkCatalogue struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec NetworkCatalogueSpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// NetworkCatalogues contains a list of network catalogs
type NetworkCatalogues struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetworkCatalogue `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetworkCatalogue{}, &NetworkCatalogues{})
}
