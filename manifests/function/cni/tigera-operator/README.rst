Calico Operator
===============

This function contains a Kubernetes operator which manages the lifecycle of a Calico or Calico Enterprise installation on Kubernetes.

* Tigera Calico operator and custom resource definitions are defined in `tigera-operator.yaml`_.
* Calico is installed via tigera operator using `custom-resources.yaml`_ by specifying the default variant of product to be installed.
* The `tigera-operator.yaml`_ is taken from `GitHub URL`_.
* Included Operator version: v1.13.0
* Included Calico version: v3.17.0

To know more about tigera installation see the `installation reference`_.

.. _tigera-operator.yaml: https://github.com/airshipit/airshipctl/tree/master/manifests/function/tigera-operator/v1.13.0/upstream/tigera-operator.yaml
.. _custom-resources.yaml: https://github.com/airshipit/airshipctl/tree/master/manifests/function/tigera-operator/v1.13.0/custom-resources.yaml
.. _GitHub URL: https://docs.projectcalico.org/manifests/tigera-operator.yaml
.. _installation reference: https://docs.projectcalico.org/getting-started/kubernetes/quickstart
