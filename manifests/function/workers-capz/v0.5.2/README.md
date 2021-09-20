# CAPZ Worker Resources
This folder contains the manifests necessary to deploy target cluster on Azure cloud.
These manifests were generated using **clusterctl generate** command with **public flavor** and then broken down into three manifests:
- cluster.yaml - this manifest is located in ../../k8scontrol-capz folder.
- controlplane.yaml - this manifest is located in ../../k8scontrol-capz folder.
- workers.yaml - provides the manifests for MachineDeployment, AzureMachineTemplate, and KubeadmConfigTemplate resources.