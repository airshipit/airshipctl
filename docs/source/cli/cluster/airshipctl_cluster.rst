.. _airshipctl_cluster:

airshipctl cluster
------------------

Airshipctl command to manage kubernetes clusters

Synopsis
~~~~~~~~


Provides capabilities for interacting with a Kubernetes cluster,
such as getting status and deploying initial infrastructure.


Options
~~~~~~~

::

  -h, --help   help for cluster

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl <airshipctl>` 	 - A unified command line tool for management of end-to-end kubernetes cluster deployment on cloud infrastructure environments.
* :ref:`airshipctl cluster get-kubeconfig <airshipctl_cluster_get-kubeconfig>` 	 - Airshipctl command to retrieve kubeconfig for a desired cluster
* :ref:`airshipctl cluster list <airshipctl_cluster_list>` 	 - Airshipctl command to get and list defined clusters
* :ref:`airshipctl cluster rotate-sa-token <airshipctl_cluster_rotate-sa-token>` 	 - Airshipctl command to rotate tokens of Service Account(s)
* :ref:`airshipctl cluster status <airshipctl_cluster_status>` 	 - Retrieve statuses of deployed cluster components

