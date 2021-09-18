.. _airshipctl_cluster_get-kubeconfig:

airshipctl cluster get-kubeconfig
---------------------------------

Airshipctl command to retrieve kubeconfig for a desired cluster(s)

Synopsis
~~~~~~~~


Retrieves kubeconfig of the cluster(s) and prints it to stdout.

If you specify single CLUSTER_NAME, kubeconfig will have a CurrentContext set to CLUSTER_NAME and
will have its context defined.

If you specify multiple CLUSTER_NAME args, kubeconfig will contain contexts for all of them, but current one
won't be specified.

If you don't specify CLUSTER_NAME, kubeconfig will have multiple contexts for every cluster
in the airship site. Context names will correspond to cluster names. CurrentContext will be empty.


::

  airshipctl cluster get-kubeconfig [CLUSTER_NAME...] [flags]

Examples
~~~~~~~~

::


  Retrieve target-cluster kubeconfig
  # airshipctl cluster get-kubeconfig target-cluster

  Retrieve kubeconfig for the entire site; the kubeconfig will have context for every cluster
  # airshipctl cluster get-kubeconfig

  Specify a file where kubeconfig should be written
  # airshipctl cluster get-kubeconfig --file ~/my-kubeconfig

  Merge site kubeconfig with existing kubeconfig file.
  Keep in mind that this can override a context if it has the same name
  Airshipctl will overwrite the contents of the file, if you want merge with existing file, specify "--merge" flag
  # airshipctl cluster get-kubeconfig --file ~/.airship/kubeconfig --merge


Options
~~~~~~~

::

  -f, --file string   specify where to write kubeconfig file. If flag isn't specified, airshipctl will write it to stdout
  -h, --help          help for get-kubeconfig
      --merge         specify if you want to merge kubeconfig with the one that exists at --file location

Options inherited from parent commands
~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~~

::

      --airshipconf string   path to the airshipctl configuration file. Defaults to "$HOME/.airship/config"
      --debug                enable verbose output

SEE ALSO
~~~~~~~~

* :ref:`airshipctl cluster <airshipctl_cluster>` 	 - Airshipctl command to manage kubernetes clusters

