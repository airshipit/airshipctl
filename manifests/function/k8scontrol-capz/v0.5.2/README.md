# CAPZ Cluster and Control Plane Resources
This folder contains the manifests necessary to deploy target cluster on Azure cloud.
These manifests were generated using **clusterctl generate** command with **public flavor** and then broken down into three manifests:
- cluster.yaml - provides the generic Cluster, AzureCluster, AzureClusterIdentity, and Secret (for client ID) resources.
- controlplane.yaml - provides the KubeadmControlPlane and AzureMachineTemplate resources.
- workers.yaml - this manifest is located in ../../workers-capz folder