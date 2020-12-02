# Calico for Azure Target Cluster

Azure does not currently support Calico networking. The reason is Azure does not allow traffic with unknown source IPs.
As a workaround, it is recommended that Azure clusters use the Calico spec below that uses VXLAN.

```bash
https://raw.githubusercontent.com/kubernetes-sigs/cluster-api-provider-azure/master/templates/addons/calico.yaml
```

You can find more about Calico on Azure [here](https://docs.projectcalico.org/reference/public-cloud/azure).
