# Zuul Gate Scripts for Bootstrap Container/Ephemeral Cluster/Azure Target Cluster
This folder container the Zuul Gate Scripts for configuring the tools necessary to
automatically validate the building of Bootstrap Container (Go app + Docker Image),
deployment of Ephemeral Cluster on Azure Cloud and Google Cloud, then using the
Azure Ephemeral Cluster to deploy the Target Cluster on Azure Cloud.

## Installing and Configuring Tools
The following scripts are used to install and configure tools such as "kubectl", "Go", "Kustomize" and "airshipctl":
- 01_install_kubectl.sh - installs "kubectl" CLI
- 02_install_go.sh - installs the "Go" language
- 03_install_kustomize_docker.sh - install "kustomize" CLI
- 21_systemwide_executable.sh - build the "airshipctl" CLI

## Bootstrap Container and Ephemeral Cluster
The following scrips are used to deploy the Ephemeral cluster on Azure and Google Cloud.
- 41_deploy_azure_ephemeral_cluster.sh - creates the Azure Bootstrap container that deploys the Azure (AKS) Ephemeral cluster
- 41_initialize_management_cluster.sh - creates the GCP Bootstrap container that deploys the GCP (GKE) Ephemeral cluster

> NOTE: the Bootstrap container images shall be built and pushed to **quay.io** registry prior to executing these scripts.
## Initializing the Ephemeral cluster and Deploying the Target Cluster
The following scripts initialize the Ephemeral cluster with CAPI and CAPZ components
and deploy the Target/Workload cluster on the Azure Cloud platform.
- 41_initialize_management_cluster.sh - initializes the Azure Ephemeral cluster with CAPI and CAPZ components
- 51_deploy_workload_cluster.sh - deploys a Target/Workload cluster on the Azure Cloud platform

And last but not least, the following scripts is a clean up script, deleting all resources created
the public clouds, including the ephemeral clusters.
- 100_clean_up_resources.sh

## Supporting Local Test Scripts
The scripts in this section are used for testing the end-to-end testing pipeline outside the Zuul
environment. It simulates the Zuul pipeline on a clean remote VM, e.g., VM created on Azure Cloud.
- 201_zuul_local_test.sh - simulates the sequence of scripts to run on a Zuul environment.
- 200_transfer_airshipctl.sh - this script transfers the airshipctl local repository to the test VM then executes 201_zuul_local_test.sh
- 200_configure_test_vm.sh - Prepares the test VM with basic tools such as "make" and "docker", then executes 200_transfer_airshipctl.sh.

By executing *200_configure_test_vm.sh* on a development server will trigger the entire test pipeline, i.e., "Zero Touch" local test.

Pre-requisite: the *200_configure_test_vm.sh* requires a special script file that exports environment variables specific for the
Azure and GCP Cloud account credentials. See template for this script below:

```bash
# Azure cloud authentication credentials.
export AZURE_SUBSCRIPTION_ID="<Your Azure Subscription ID>"
export AZURE_TENANT_ID="<Your Tenant ID>"
export AZURE_CLIENT_ID="<Your Service Principal ID>"
export AZURE_CLIENT_SECRET="<Your Service Principal Secret>"

# To use the default public cloud, otherwise set to AzureChinaCloud|AzureGermanCloud|AzureUSGovernmentCloud
export AZURE_ENVIRONMENT="AzurePublicCloud"

export AZURE_SUBSCRIPTION_ID_B64="$(echo -n "$AZURE_SUBSCRIPTION_ID" | base64 | tr -d '\n')"
export AZURE_TENANT_ID_B64="$(echo -n "$AZURE_TENANT_ID" | base64 | tr -d '\n')"
export AZURE_CLIENT_ID_B64="$(echo -n "$AZURE_CLIENT_ID" | base64 | tr -d '\n')"
export AZURE_CLIENT_SECRET_B64="$(echo -n "$AZURE_CLIENT_SECRET" | base64 | tr -d '\n')"

# GCP Environment Variables
export GCP_PROJECT=<Your Google Cloud Project ID>
export GCP_ACCOUNT=<Your Google Cloud Account ID>
```
