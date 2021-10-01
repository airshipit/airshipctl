# Reference Site for Deploying Public facing Target Cluster on Azure Cloud
This site provides the manifests to deploy a Target cluster on Azure cloud,
that is accessible from the public Internet.

## Pre-Conditions
You will need to provide the Azure cloud (Service Pricipal role Contributor) credentials in the imported secrets.
See *manifests/site/reference-az-public-cluster/target/encrypted/results/imported/secrets.yaml*

You have to edit this file with the *sops* CLI and provide the following credential data:
* subscriptionID - enter value as is
* tenantID - enter value as is
* clientID - enter value as is
* clientSecret - base64 encoded Client Secret

```yaml
apiVersion: airshipit.org/v1alpha1
kind: VariableCatalogue
metadata:
    labels:
        airshipit.org/deploy-k8s: "false"
    name: imported-secrets
azure:
    identity:
        subscriptionID: <your Azure Subscription ID>
        tenantID: <your Azure Subscription's Tenant ID>
        clientID: <your Azure Subscription's Client ID>
        clientSecret: <your Azure Subscription's Client Secret - base64>
```

## Deploying Your Target Cluster on Azure Cloud

First you need to deploy an ephemeral cluster with Kind.

>IMPORTANT: You need to delete all references to the **target-cluster** in $HOME/.airship/kubeconfig otherwise it will not work.
>
>Easy to delete $HOME/.airship/kubeconfig file before creating the ephemeral cluster.


```sh
CLUSTER=ephemeral-cluster <path to your airshipctl repo>/tools/deployment/kind/start_kind.sh
```

Once your ephemeral cluster has been created you can start the deployment as follow:

```sh
airshipctl plan run deploy-gating --debug
```

After a few minutes your cluster should be up and operational.
To check you can go to https://portal.azure.com/ and verify that control plane and worker VMs
have been created.

## Multi-tenancy
The CAPZ V0.5.0 supports proprietary Multitenancy, meaning that you can create multiple Target clusters
using different Azure subscriptions. This is achieved through the resources AzureCluster (subscriptionID),
AzureClusterIdentity (tenant ID, client ID) and Secret (client secret).

In this reference site, these credentials data is provided in an (sops) encrypted file (see Pre Conditions section above),
which is used to patch the Azure account credentials to the resource mentioned in this section.

## Validating the Clusterctl Move
In order to verify that CAPI/CAPZ Management components moved correctly to the Target cluster you can try to scale the
number of nodes up and down and see if the number of nodes increase and decrease as specified.

A more elaborated test would be to deploy multiple Pods, ideally the replica count for a Deployment to be higher than the
number of worker nodes. Scale down the number of worker nodes and verify that the Pods are redistributed among remaining nodes.

## Troubleshooting
You will find some tips for troubleshooting [here](https://capz.sigs.k8s.io/topics/troubleshooting.html)
