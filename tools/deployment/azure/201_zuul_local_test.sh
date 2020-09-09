#!/bin/bash

# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This script simulates the Zuul gates for validation of Azure cloud integration.
# It goes through all gates for locally in the remote VM.

echo "----- Using default airship directory"
export AIRSHIPDIR="/home/zuul/src/opendev.org/airship/airshipctl"
export AIRSHIPCTL_WS="/home/zuul/src/opendev.org/airship/airshipctl"
export PATH=$PATH:/usr/local/go/bin/

# Setting Public Cloud Credentials as Environment Variables
echo "source ~/.profile"
source ~/.profile

echo "Azure Subscription ID = $AZURE_SUBSCRIPTION_ID"
echo "Azure Tenant ID = $AZURE_TENANT_ID"
echo "Google Cloud Project ID = $GCP_PROJECT"
echo "Google Cloud Account ID = $GCP_ACCOUNT"

cd $AIRSHIPCTL_WS
echo "----- Airship Directory = $AIRSHIPCTL_WS"

# Installation of Kubectl
echo "************************************************************************"
echo "***** Installation of Kubectl ..."
./tools/deployment/01_install_kubectl.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to Install Kubectl CLI"
    exit 1
fi

# Build Kind Cluster
echo "************************************************************************"
echo "***** Building Kind Cluster ..."
./tools/deployment/azure/11_build_kind_cluster.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to build Kind cluster"
    exit 1
fi

# Building airshipctl command
echo "************************************************************************"
echo "***** Building airshipctl command ..."
./tools/deployment/21_systemwide_executable.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to build airshipctl CLI"
    exit 1
fi

# Creating Airship config file
echo "************************************************************************"
echo "***** Creating Airship config file ..."
./tools/deployment/azure/31_create_configs.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to create airshipctl config file"
    exit 1
fi

# Initializing CAPI and CAPZ components for the Managemeng cluster
echo "************************************************************************"
echo "***** Initializing CAPI and CAPZ components for the Managemeng cluster ..."
./tools/deployment/azure/41_initialize_management_cluster.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to initialize the Ephemeral cluster with CAPI/CAPZ components"
    exit 1
fi

# Deploying the Target Cluster in Azure cloud
echo "************************************************************************"
echo "***** Deploying the Target Cluster in azure cloud ..."
./tools/deployment/azure/51_deploy_workload_cluster.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to deploy Target/Workload cluster on Azure Cloud"
    exit 1
fi

# Sleep for 15 min before start cleaning up everything.
echo "Waiting for 15 min..."
sleep 15m

# Cleaning up Resources
echo "************************************************************************"
echo "***** Cleaning up resources ..."
./tools/deployment/azure/100_clean_up_resources.sh
if [ $? -ne 0 ]; then
    echo ">>>>> Failed to clean up all public cloud resources created to this test"
    exit 1
fi
