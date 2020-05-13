Airshipctl
==========

Airshipctl is a command-line interface that enables users to manage declarative
infrastructure and software.

Airshipctl aims to provide a seamless experience for operators wishing to
leverage the best open source options such as the `Cluster API`_, `Metal
Kubed`_, Kustomize_, and kubeadm_ by providing a straight forward and easily
approachable interface.

This project is the heart of our effort to produce Airship 2.0, which has
three main evolutions from `Airship 1.0`_:

* Expand our use of entrenched upstream projects.
* Embrace Kubernetes Custom Resource Definitions (CRD) – everything becomes an
  object in Kubernetes.
* Make the Airship control plane ephemeral.

To learn more about the Airship 2.0 evolution, reference the
`Airship blog series`_.

Contributing
------------

Airshipctl is under active development and welcomes new developers! Please
read our `developer guide`_ to begin contributing.

We also encourage new contributors and operators alike to join us in our
`Slack workspace`_ and subscribe to our `mailing lists`_.

You can learn more about Airship on the `Airship wiki`_.

.. _Airship 1.0: https://docs.airshipit.org/treasuremap
.. _Airship blog series: https://www.airshipit.org/blog/airship-blog-series-1-evolution-towards-2.0
.. _Airship wiki: https://wiki.openstack.org/wiki/Airship
.. _Cluster API: https://github.com/kubernetes-sigs/cluster-api
.. _developer guide: https://docs.airshipit.org/airshipctl/developers.html
.. _kubeadm: https://kubernetes.io/docs/reference/setup-tools/kubeadm/kubeadm
.. _Kustomize: https://github.com/kubernetes-sigs/kustomize
.. _mailing lists: http://lists.airshipit.org/cgi-bin/mailman/listinfo
.. _Metal Kubed: https://metal3.io
.. _Slack workspace: http://airshipit.org/slack
